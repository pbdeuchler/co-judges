#!/usr/bin/env bash
# Build the site and publish it to S3 behind CloudFront at co-judges.philip.computer.
# Infrastructure lives in ../terraform/prod (client_sites "co-judges" in cloudfront.tf).
#
# Usage: ./deploy.sh [--scrape] [--dry-run]
#   --scrape   re-fetch and re-parse the state's reports before building
#   --dry-run  show what would upload/delete; skip the CloudFront invalidation
#
# Needs: go, uv (with --scrape), aws CLI credentials that can write the bucket and invalidate.
set -euo pipefail
cd "$(dirname "$0")"

BUCKET="${BUCKET:-co-judges-client}"
DOMAIN="${DOMAIN:-co-judges.philip.computer}"
scrape=0
dry=()
for arg in "$@"; do
  case "$arg" in
    --scrape) scrape=1 ;;
    --dry-run) dry=(--dryrun) ;;
    *) echo "usage: $0 [--scrape] [--dry-run]" >&2; exit 2 ;;
  esac
done

if ((scrape)); then
  uv run --script scrape/fetch.py
  uv run --script scrape/parse.py
fi
go test ./gen
go run ./gen -data data/evaluations.json -out public

sync() { aws s3 sync public "s3://$BUCKET" --only-show-errors ${dry[@]+"${dry[@]}"} "$@"; }

# Fonts never change in place; stylesheet, script and icon are not fingerprinted, so stay short;
# pages revalidate quickly so new data shows within minutes even without an invalidation.
sync --exclude '*' --include 'fonts/*' --content-type font/woff2 --cache-control 'public,max-age=31536000,immutable'
sync --exclude '*' --include '*.css' --include '*.js' --include '*.svg' --cache-control 'public,max-age=3600'
sync --exclude 'fonts/*' --exclude '*.css' --exclude '*.js' --exclude '*.svg' --cache-control 'public,max-age=300'
# Remove objects that no longer exist locally (uploads nothing new: everything is already synced).
sync --delete

if ((${#dry[@]})); then
  echo "dry run: skipped invalidation"
  exit 0
fi

dist=$(aws cloudfront list-distributions --output text \
  --query "DistributionList.Items[?Aliases.Items && contains(Aliases.Items, '$DOMAIN')].Id | [0]")
if [[ -z "$dist" || "$dist" == "None" ]]; then
  echo "no CloudFront distribution has alias $DOMAIN; apply ../terraform/prod first" >&2
  exit 1
fi
aws cloudfront create-invalidation --distribution-id "$dist" --paths '/*' --query 'Invalidation.Id' --output text >/dev/null
echo "deployed to https://$DOMAIN (invalidation queued on $dist)"
