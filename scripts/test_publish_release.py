import copy
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from publish_release import ASSETS, GitHub, digest, publish_and_prune, release_date


def release(tag, **kwargs):
    return dict(tag_name=tag, draft=False, prerelease=False, published_at="2026-10-20T00:00:00Z", assets=[], **kwargs)


class FakeGitHub(GitHub):
    def __init__(self, root, existing=()):
        super().__init__("fixture/repo")
        self.root = root
        self.releases = {r["tag_name"]: copy.deepcopy(r) for r in existing}
        self.tags = set(self.releases)
        self.calls = []
        self.fail = None
        self.bad_digest = False
        self.bad_size = False
        self.missing_asset = False
        self.no_digest = False
        self.fail_tag_once = False

    def release(self, tag):
        return copy.deepcopy(self.releases.get(tag))

    def tag(self, tag):
        return {"ref": tag} if tag in self.tags else None

    def pages(self, endpoint):
        self.calls.append(("pages", endpoint))
        if endpoint == "releases":
            return copy.deepcopy(list(self.releases.values()))
        return [{"name": tag} for tag in sorted(self.tags)]

    def run(self, *args):
        self.calls.append(args)
        if self.fail and args[:2] == self.fail:
            raise subprocess.CalledProcessError(1, args, stderr="HTTP 500")
        if args[:2] in (("release", "create"), ("release", "upload")):
            tag = args[2]
            item = self.releases.setdefault(tag, release(tag))
            if args[1] == "create":
                assert "--draft" in args
                item["draft"] = True
                item["published_at"] = None
            item["assets"] = [{"name": name, "size": (self.root / name).stat().st_size,
                               "state": "uploaded", "digest": digest(self.root / name)} for name in ASSETS]
            if self.bad_digest:
                item["assets"][0]["digest"] = "sha256:wrong"
            if self.bad_size:
                item["assets"][0]["size"] += 1
            if self.missing_asset:
                item["assets"].pop()
            if self.no_digest:
                for asset in item["assets"]:
                    asset["digest"] = None
            self.tags.add(tag)
        elif args[:2] == ("release", "edit"):
            item = self.releases[args[2]]
            item["draft"] = False
            item["published_at"] = "2026-10-23T00:00:00Z"
        elif args[:2] == ("release", "delete"):
            tag = args[2]
            self.releases.pop(tag, None)
            if self.fail_tag_once:
                self.fail_tag_once = False
                raise subprocess.CalledProcessError(1, args, stderr="tag cleanup failed")
            self.tags.discard(tag)
        elif args[:3] == ("api", "--method", "DELETE"):
            self.tags.discard(args[3].rsplit("/", 1)[1])
        elif args[:2] == ("release", "download"):
            name = args[args.index("--pattern") + 1]
            directory = args[args.index("--dir") + 1]
            (Path(directory) / name).write_bytes((self.root / name).read_bytes())
        else:
            raise AssertionError(f"Unexpected GitHub command: {args}")
        return ""


class ReleaseTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        for name in ASSETS:
            (self.root / name).write_bytes((name + " fixture").encode())

    def publish(self, client):
        publish_and_prune(client, self.root, "geoip-20261023", "commit-sha")

    def test_publication_prunes_only_older_dated_releases_and_tags(self):
        protected = [release("geoip-20261024"), release("v1.0"), release("geoip-20260230")]
        draft = release("geoip-20261019"); draft["draft"] = True
        pre = release("geoip-20261018"); pre["prerelease"] = True
        protected += [draft, pre]
        client = FakeGitHub(self.root, [release("geoip-20261020"), *protected])
        self.publish(client)
        self.assertEqual(set(client.releases), {"geoip-20261023", *(r["tag_name"] for r in protected)})
        self.assertEqual(client.tags, set(client.releases))
        actions = [c[:2] for c in client.calls]
        self.assertLess(actions.index(("release", "edit")), actions.index(("release", "delete")))
        self.assertIn("--cleanup-tag", next(c for c in client.calls if c[:2] == ("release", "delete")))

    def test_same_day_and_draft_retry_reuse_release(self):
        for draft in (False, True):
            with self.subTest(draft=draft):
                existing = release("geoip-20261023"); existing["draft"] = draft
                client = FakeGitHub(self.root, [existing])
                self.publish(client)
                self.publish(client)
                self.assertFalse(any(c[:2] == ("release", "create") for c in client.calls))
                self.assertIn("--clobber", next(c for c in client.calls if c[:2] == ("release", "upload")))
                self.assertEqual(client.tags, {"geoip-20261023"})

    def test_failed_upload_or_publication_preserves_old_release(self):
        for operation in ("create", "upload", "edit"):
            with self.subTest(operation=operation):
                old = [release("geoip-20261020")]
                if operation == "upload": old.append(release("geoip-20261023"))
                client = FakeGitHub(self.root, old); client.fail = ("release", operation)
                with self.assertRaises(subprocess.CalledProcessError): self.publish(client)
                self.assertIn("geoip-20261020", client.releases)
                self.assertFalse(any(c[:2] == ("pages", "releases") for c in client.calls))

    def test_failed_asset_verification_never_publishes_or_cleans(self):
        for failure in ("bad_digest", "bad_size", "missing_asset"):
            with self.subTest(failure=failure):
                client = FakeGitHub(self.root, [release("geoip-20261020")]); setattr(client, failure, True)
                with self.assertRaises(RuntimeError): self.publish(client)
                self.assertIn("geoip-20261020", client.releases)
                self.assertTrue(client.releases["geoip-20261023"]["draft"])
                self.assertFalse(any(c[:2] == ("release", "edit") for c in client.calls))

    def test_missing_generation_asset_has_no_remote_mutations(self):
        (self.root / ASSETS[0]).unlink()
        client = FakeGitHub(self.root)
        with self.assertRaises(RuntimeError): self.publish(client)
        self.assertEqual(client.calls, [])

    def test_no_server_digest_downloads_and_hashes_assets(self):
        client = FakeGitHub(self.root); client.no_digest = True
        self.publish(client)
        self.assertEqual(sum(c[:2] == ("release", "download") for c in client.calls), 2 * len(ASSETS))

    def test_partial_deletion_and_orphan_tags_are_retried(self):
        client = FakeGitHub(self.root, [release("geoip-20261020")]); client.fail_tag_once = True
        client.tags.add("geoip-20261017")
        client.tags.add("v0.1")
        self.publish(client)
        self.assertEqual(client.tags, {"geoip-20261023", "v0.1"})

    def test_delete_failure_fails_run_and_next_attempt_recovers(self):
        client = FakeGitHub(self.root, [release("geoip-20261020")]); client.fail = ("release", "delete")
        with self.assertRaises(subprocess.CalledProcessError): self.publish(client)
        self.assertIn("geoip-20261020", client.releases)
        client.fail = None
        self.publish(client)
        self.assertEqual(set(client.releases), {"geoip-20261023"})

    def test_pagination_and_http_errors(self):
        client = GitHub("fixture/repo")
        with patch.object(client, "run", return_value=json.dumps([[{"tag_name":"one"}], [{"tag_name":"two"}]])) as run:
            self.assertEqual(len(client.pages("releases")), 2)
            self.assertIn("--paginate", run.call_args.args)
            self.assertIn("--slurp", run.call_args.args)
        with patch.object(client, "run", side_effect=subprocess.CalledProcessError(1, "gh", stderr="HTTP 404")):
            self.assertIsNone(client.release("missing"))
        with patch.object(client, "run", side_effect=subprocess.CalledProcessError(1, "gh", stderr="HTTP 403")):
            with self.assertRaises(subprocess.CalledProcessError): client.release("denied")

    def test_invalid_tag_and_protected_same_day_prerelease(self):
        self.assertIsNone(release_date("geoip-20260230"))
        self.assertIsNone(release_date("v1.0"))
        item = release("geoip-20261023"); item["prerelease"] = True
        client = FakeGitHub(self.root, [item])
        with self.assertRaises(RuntimeError): self.publish(client)
        self.assertEqual(client.calls, [])

    def test_already_removed_release_is_idempotent(self):
        client = GitHub("fixture/repo")
        with patch.object(client, "run", side_effect=subprocess.CalledProcessError(1, "gh", stderr="HTTP 404")), \
             patch.object(client, "release", return_value=None), \
             patch.object(client, "tag", return_value=None):
            client.delete_release("geoip-20261020")


if __name__ == "__main__":
    unittest.main()
