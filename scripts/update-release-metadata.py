#!/usr/bin/env python3
import json
import sys
import boto3
import re


def fetch_metadata():
  '''
  Pull metadata from s3 as byte stream.
  '''
  client = boto3.client('s3')
  try:
    resp = client.get_object(
        Bucket='release-metadata',
        Key='release-versions'
    )
  except client.exceptions.NoSuchKey:
    sys.stderr.write('No release-metadata key found.  Check permissions.\n')
    return 1

  return resp['Body'].read()


def write_metadata(data):
  '''
  Write metadata file from byte stream.
  '''
  client = boto3.client('s3')
  resp = client.put_object(
      Body=data,
      Bucket='release-metadata',
      Key='release-versions'
  )

  if resp.ResponseMetadata.HTTPStatusCode != 200:
    sys.stderr.write('Unable to upload file.  Dumping response metadata.\n')
    print(resp, file=sys.stderr)
    return 1


def main(module_name, new_ref):
  if not new_ref.startswith('refs/tags/'):
    sys.stderr.write(
        f'Unknown reference format {new_ref}.  Expecting refs/tags/v<version>\n')
    return 1
  # extract the version
  new_tag = re.sub('[A-Za-z/]', '', new_ref)

  print(f'Releasing version {new_tag} of module {module_name}')

  release_data = fetch_metadata()
  data = json.loads(release_data)
  module_versions = data.get(module_name)
  if not module_versions:
    sys.stderr.write(
        f'Module {module_name} does not exist in release metadata.\n')
    return 1
  if new_tag in module_versions:
    print(f"Tag {new_tag} is already released for module {module_name}. Exiting.")
    return 0
  module_versions.insert(0, new_tag)
  write_metadata(json.dumps(data).encode('utf-8'))


def usage():
  print("Usage: python3 update-release-metadata.py <module> <ref>")
  return 2


if __name__ == "__main__":
  if len(sys.argv) != 3:
    sys.exit(usage())
  sys.exit(main(*sys.argv[1:]))
