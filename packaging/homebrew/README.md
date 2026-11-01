# Publish to Homebrew

The checked-in formula is a template for `ContractorKeith/homebrew-tap`.
`REPLACE_AFTER_RELEASE` hashes deliberately prevent installation before assets exist.

1. After macOS/Linux skeptic review and triage, tag the approved commit `v0.1.0`
   and push the tag. Wait for the release workflow to publish all four tarballs
   and `checksums.txt` on the GitHub Release.
2. Run `bash packaging/homebrew/update-formula.sh v0.1.0`. It downloads the
   checksums, validates all four entries, and rewrites the version and sha256
   lines in `packaging/homebrew/mdkb.rb`. Review the diff, copy it to
   `Formula/mdkb.rb` in a checkout of the tap, and open a PR there. After merge,
   verify `brew install contractorkeith/tap/mdkb` and `brew test mdkb` on macOS
   and Linux.

The script never tags, publishes, or pushes to the tap. For offline testing,
pass a checksums file and optional output formula path as its second/third arguments.
Archives include the `v` prefix: `mdkb_v0.1.0_darwin_arm64.tar.gz`.
