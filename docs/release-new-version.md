# Guide for Release new version

Just create a new tag and push, the GitHub Actions will to the rest.

```bash
git tag -v v2.5.1
git push origin v2.5.1
```

After the GitHub Actions finished, the new version will be released to GitHub Releases.

https://github.com/gobackup/gobackup/releases

Edit the latest release, and add the release note.

## Release to Homebrew

The GitHub Actions also have a step to release to Homebrew.

But there may have a problem, the Homebrew Core repo may not update the formula.

So it need visit: https://github.com/huacnlee/homebrew-core

And click `Sync Fork` button, then click `Update Branch` button to update the formula.

Then retry the Homebrew release GitHub Action.
