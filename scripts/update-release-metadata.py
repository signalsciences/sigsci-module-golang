#!/usr/bin/env python3
import json
import sys
import boto3


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
  except boto3.S3.Client.exceptions.NoSuchKey:
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


def main(module_name, new_tag):
  release_data = fetch_metadata()
  data = json.loads(release_data)
  module_versions = data[module_name]
  if new_tag in module_versions:
    print(f"Tag {new_tag} is already released for module {module_name}. Exiting.")
    return 0
  module_versions.insert(0, new_tag)
  write_metadata(json.dumps(data).encode('utf-8'))


def usage():
  print("Usage: python3 update-release-metadata.py <module> <version>")
  return 2


if __name__ == "__main__":
  if len(sys.argv) != 3:
    sys.exit(usage())
  sys.exit(main(*sys.argv[1:]))
