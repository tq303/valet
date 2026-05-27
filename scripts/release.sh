#!/bin/bash
set -e

CURRENT=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
NEXT=$(echo $CURRENT | awk -F. '{print $1"."$2"."$3+1}')

echo "Releasing $NEXT"

git tag $NEXT
git push origin $NEXT

cd npm
npm version ${NEXT#v} --no-git-tag-version
npm publish --access=public
cd ..

git add npm/package.json
git commit -m "chore: release $NEXT"
git push
