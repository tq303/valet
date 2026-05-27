#!/bin/bash
set -e

CURRENT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
NEXT_TAG=$(echo $CURRENT_TAG | awk -F. '{print $1"."$2"."$3+1}')

echo "Releasing ${NEXT_TAG}"

git tag ${NEXT_TAG}
git push origin ${NEXT_TAG}

cd npm
npm version ${NEXT_TAG#v} --no-git-tag-version
npm publish --access=public
cd ..

git add npm/package.json
git commit -m "release: npm ${NEXT_TAG}"
git push
