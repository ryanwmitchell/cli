#!/bin/sh

# Adapted from https://github.com/leopardslab/dunner/blob/master/release/publish_rpm_to_bintray.sh

set -e

SUBJECT="dopplerhq"
REPO="doppler-rpm"
PACKAGE="doppler"

if [ -z "$BINTRAY_USER" ]; then
  echo "BINTRAY_USER is not set"
  exit 1
fi

if [ -z "$BINTRAY_API_KEY" ]; then
  echo "BINTRAY_API_KEY is not set"
  exit 1
fi

getVersion () {
  # use --abbrev=0 to ensure the returned tag was explicitly set
  VERSION=$(git describe --abbrev=0);
}

listRpmArtifacts() {
  FILES=$(find dist/*.rpm  -type f)
}

bintrayCreateVersion () {
  URL="https://api.bintray.com/packages/$SUBJECT/$REPO/$PACKAGE/versions"
  BODY="{ \"name\": \"$VERSION\", \"github_use_tag_release_notes\": false, \"vcs_tag\": \"$VERSION\" }"
  echo "Creating rpm package version $VERSION"
  RESPONSE_CODE=$(curl -X POST -d "$BODY" -H "Content-Type: application/json" -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -s -w "%{http_code}" -o /dev/null);
  if [[ "$(echo $RESPONSE_CODE | head -c2)" != "20" ]]; then
    echo "Unable to create package version, HTTP response code: $RESPONSE_CODE"
    exit 1
  fi
}

bintrayUseGitHubReleaseNotes () {
  URL="https://api.bintray.com/packages/$SUBJECT/$REPO/$PACKAGE/versions/$VERSION"
  BODY="{ \"vcs_tag\": \"v0.0.33\", \"github_use_tag_release_notes\": \"true\" }"
  echo "Using release notes from GitHub"
  RESPONSE_CODE=$(curl -X PATCH -d "$BODY" -H "Content-Type: application/json" -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -s -w "%{http_code}" -o /dev/null);
  if [[ "$(echo $RESPONSE_CODE | head -c2)" != "20" ]]; then
    echo "Unable to create package version, HTTP response code: $RESPONSE_CODE"
    exit 1
  fi
}

bintrayUpload () {
  for i in $FILES; do
    FILENAME=${i##*/}
    ARCH=$(echo ${FILENAME##*_} | cut -d '.' -f 1)
    URL="https://api.bintray.com/content/$SUBJECT/$REPO/$PACKAGE/$VERSION/$ARCH/$FILENAME?publish=1&override=1"

    echo "Uploading $URL"
    RESPONSE_CODE=$(curl -T $i -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -I -s -w "%{http_code}" -o /dev/null);
    if [[ "$(echo $RESPONSE_CODE | head -c2)" != "20" ]]; then
      echo "Unable to upload, HTTP response code: $RESPONSE_CODE"
      exit 1
    fi
  done;
}

bintraySetDownloads () {
  for i in $FILES; do
    FILENAME=${i##*/}
    ARCH=$(echo ${FILENAME##*_} | cut -d '.' -f 1)
    URL="https://api.bintray.com/file_metadata/$SUBJECT/$REPO/$ARCH/$FILENAME"

    echo "Putting $FILENAME in $PACKAGE's download list"
    RESPONSE_CODE=$(curl -X PUT -d "{ \"list_in_downloads\": true }" -H "Content-Type: application/json" -u$BINTRAY_USER:$BINTRAY_API_KEY $URL -s -w "%{http_code}" -o /dev/null);

    if [ "$(echo $RESPONSE_CODE | head -c2)" != "20" ]; then
        echo "Unable to put in download list, HTTP response code: $RESPONSE_CODE"
        echo "URL: $URL"
        exit 1
    fi
  done
}

snooze () {
    echo "\nSleeping for 30 seconds. Have a coffee..."
    sleep 30s;
    echo "Done sleeping\n"
}

printMeta () {
    echo "Publishing: $PACKAGE"
    echo "Version to be uploaded: $VERSION"
}

cleanArtifacts () {

  rm -f "$(pwd)/*.rpm"
}

cleanArtifacts
listRpmArtifacts
getVersion
printMeta
bintrayCreateVersion
bintrayUseGitHubReleaseNotes
bintrayUpload
snooze
bintraySetDownloads
