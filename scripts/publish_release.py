"""Publish verified GeoIP assets, then prune older dated releases and tags."""

import datetime
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
from urllib.parse import quote


ASSETS = (
    "geoip.dat", "country.mmdb", "country-csv.zip", "source-csv.zip",
    "sing-box-rulesets.zip", "run-summary.json", "manifest.json", "SHA256SUMS",
)


def release_date(tag):
    if not re.fullmatch(r"geoip-[0-9]{8}", tag):
        return None
    try:
        return datetime.datetime.strptime(tag[6:], "%Y%m%d").date()
    except ValueError:
        return None


def digest(path):
    with path.open("rb") as stream:
        return "sha256:" + hashlib.file_digest(stream, "sha256").hexdigest()


class GitHub:
    def __init__(self, repo):
        self.repo = repo

    def run(self, *args):
        return subprocess.run(
            ["gh", *args], check=True, capture_output=True, text=True,
        ).stdout

    def api(self, endpoint):
        return json.loads(self.run("api", f"repos/{self.repo}/{endpoint}"))

    def optional_api(self, endpoint):
        try:
            return self.api(endpoint)
        except subprocess.CalledProcessError as error:
            # Authentication, rate-limit and server errors must never look absent.
            if "HTTP 404" in (error.stderr or ""):
                return None
            raise

    def release(self, tag):
        return self.optional_api(f"releases/tags/{quote(tag, safe='')}")

    def release_including_drafts(self, tag):
        # GitHub's releases/tags endpoint omits drafts, so fall back to the
        # paginated release list while a newly created release is still a draft.
        release = self.release(tag)
        if release is not None:
            return release
        return next(
            (candidate for candidate in self.pages("releases")
             if candidate.get("tag_name") == tag),
            None,
        )

    def tag(self, tag):
        return self.optional_api(f"git/ref/tags/{quote(tag, safe='')}")

    def pages(self, endpoint):
        pages = json.loads(self.run(
            "api", "--paginate", "--slurp",
            f"repos/{self.repo}/{endpoint}?per_page=100",
        ))
        return [item for page in pages for item in page]

    def delete_tag(self, tag):
        try:
            self.run("api", "--method", "DELETE",
                     f"repos/{self.repo}/git/refs/tags/{quote(tag, safe='')}")
        except subprocess.CalledProcessError:
            if self.tag(tag) is not None:
                raise

    def delete_release(self, tag):
        try:
            self.run("release", "delete", tag, "--cleanup-tag", "--yes",
                     "--repo", self.repo)
        except subprocess.CalledProcessError:
            if self.release_including_drafts(tag) is not None:
                raise
            # A prior attempt may have deleted the release but not its tag.
            if self.tag(tag) is not None:
                self.delete_tag(tag)
        if self.release_including_drafts(tag) is not None or self.tag(tag) is not None:
            raise RuntimeError(f"Cleanup did not remove release and tag {tag}")


def verify_assets(client, release, root):
    assets = release.get("assets", [])
    for name in ASSETS:
        matches = [asset for asset in assets if asset.get("name") == name]
        local = root / name
        if len(matches) != 1:
            raise RuntimeError(f"Expected one uploaded {name}, found {len(matches)}")
        asset = matches[0]
        if asset.get("state") != "uploaded" or asset.get("size") != local.stat().st_size:
            raise RuntimeError(f"Incomplete or wrong-sized upload: {name}")
        expected = digest(local)
        uploaded_digest = asset.get("digest")
        if not uploaded_digest:
            with tempfile.TemporaryDirectory(prefix="geoip-verify-") as directory:
                client.run("release", "download", release["tag_name"],
                           "--pattern", name, "--dir", directory, "--repo", client.repo)
                uploaded_digest = digest(Path(directory) / name)
        if uploaded_digest != expected:
            raise RuntimeError(f"SHA-256 mismatch: {name}")


def publish_and_prune(client, root, tag, target):
    date = release_date(tag)
    if date is None:
        raise ValueError(f"Invalid dated GeoIP release: {tag}")
    for name in ASSETS:
        if not (root / name).is_file() or (root / name).stat().st_size == 0:
            raise RuntimeError(f"Missing or empty release asset: {name}")
    existing = client.release_including_drafts(tag)
    if existing and existing.get("prerelease"):
        raise RuntimeError(f"Refusing to replace prerelease {tag}")
    paths = [str(root / name) for name in ASSETS]
    if existing:
        client.run("release", "upload", tag, *paths, "--clobber", "--repo", client.repo)
    else:
        client.run("release", "create", tag, *paths, "--draft", "--target", target,
                   "--title", tag, "--notes",
                   "Daily country GeoIP generation. See run-summary.json and SHA256SUMS for source status and integrity hashes.",
                   "--repo", client.repo)
    uploaded = client.release_including_drafts(tag)
    if uploaded is None:
        raise RuntimeError(f"Release {tag} missing after upload")
    verify_assets(client, uploaded, root)
    client.run("release", "edit", tag, "--draft=false", "--latest", "--repo", client.repo)
    published = client.release_including_drafts(tag)
    if published is None or published.get("draft") or published.get("prerelease") or not published.get("published_at"):
        raise RuntimeError(f"Release {tag} was not published")
    verify_assets(client, published, root)

    releases = client.pages("releases")
    for old in releases:
        old_tag = old["tag_name"]
        old_date = release_date(old_tag)
        if old_date is None or old_date >= date or old.get("draft") or old.get("prerelease") or not old.get("published_at"):
            continue
        # Recheck immediately before deletion, including a concurrent protection change.
        current = client.release_including_drafts(old_tag)
        if current and (current.get("draft") or current.get("prerelease") or not current.get("published_at")):
            continue
        if current:
            client.delete_release(old_tag)
        elif client.tag(old_tag) is not None:
            client.delete_tag(old_tag)
        print(f"Removed older GeoIP release and tag: {old_tag}", flush=True)

    # Resume a tag deletion that failed after release removal on a previous run.
    for ref in client.pages("tags"):
        old_tag = ref["name"]
        old_date = release_date(old_tag)
        if old_date is not None and old_date < date and client.release_including_drafts(old_tag) is None:
            client.delete_tag(old_tag)
            print(f"Removed orphaned older GeoIP tag: {old_tag}", flush=True)
    print(f"Published and verified {tag}; release retention complete", flush=True)


if __name__ == "__main__":
    release_tag = "geoip-" + datetime.datetime.now(datetime.timezone.utc).strftime("%Y%m%d")
    try:
        publish_and_prune(GitHub(os.environ["GITHUB_REPOSITORY"]), Path("dist"),
                          release_tag, os.environ["GITHUB_SHA"])
    except subprocess.CalledProcessError as error:
        print(error.stderr or str(error), file=sys.stderr)
        sys.exit(1)
