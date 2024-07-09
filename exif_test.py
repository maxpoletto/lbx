#!/Users/maxp/venv/bin/python
"""Exif test"""

import exif
import exifread
import sys

fn = sys.argv[1]

with open(fn, 'rb') as image_file:
    img = exif.Image(image_file)
    for key in dir(img):
        try:
            print(key, img[key])
        except Exception as e:
            pass

with open(fn, 'rb') as f:
    tags = exifread.process_file(f)
    for k in tags.keys():
        print(k, tags[k])
