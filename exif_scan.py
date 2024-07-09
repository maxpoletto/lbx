"""Import a directory tree of photos into a database."""

import os
import re
from typing import Any, Dict, Optional

import exifread
import psycopg2

# Database connection details
DB_NAME = "lbx"
DB_USER = "maxp"
#DB_PASSWORD = "your_db_password"
#DB_HOST = "your_db_host"
#DB_PORT = "your_db_port"

# Connect to PostgreSQL database
conn = psycopg2.connect(dbname=DB_NAME, user=DB_USER)
cur = conn.cursor()

# Create master table if it doesn't exist
cur.execute("""
CREATE TABLE IF NOT EXISTS exif_metadata (
    id SERIAL PRIMARY KEY,
    filepath TEXT,
    directory TEXT,
    datetime TEXT,
    datetime_original TEXT,
    gps_latitude NUMERIC,
    gps_latitude_ref TEXT,
    gps_longitude NUMERIC,
    gps_longitude_ref TEXT,
    make TEXT,
    model TEXT,
    lens_model TEXT,
    image_description TEXT
)
""")
conn.commit()

# Create keywords table if it doesn't exist
cur.execute("""
CREATE TABLE IF NOT EXISTS keywords (
    id SERIAL PRIMARY KEY,
    keyword TEXT,
    image_id INTEGER REFERENCES exif_metadata(id)
)
""")
conn.commit()

def extract_exif_data(filepath: str) -> Dict[str, Optional[Any]]:
    exif_data: Dict[str, Optional[Any]] = {}
    try:
        with open(filepath, 'rb') as image_file:
            tags = exifread.process_file(image_file, details=False)
            
            exif_data['datetime'] = str(tags.get('EXIF DateTimeDigitized', None))
            exif_data['datetime_original'] = str(tags.get('EXIF DateTimeOriginal', None))
            
            gps_latitude = tags.get('GPS GPSLatitude', None)
            gps_latitude_ref = tags.get('GPS GPSLatitudeRef', None)
            if gps_latitude:
                exif_data['gps_latitude'] = convert_to_degrees(gps_latitude)
                exif_data['gps_latitude_ref'] = str(gps_latitude_ref) if gps_latitude_ref else None
            
            gps_longitude = tags.get('GPS GPSLongitude', None)
            gps_longitude_ref = tags.get('GPS GPSLongitudeRef', None)
            if gps_longitude:
                exif_data['gps_longitude'] = convert_to_degrees(gps_longitude)
                exif_data['gps_longitude_ref'] = str(gps_longitude_ref) if gps_longitude_ref else None
            
            exif_data['make'] = str(tags.get('Image Make', None))
            exif_data['model'] = str(tags.get('Image Model', None))
            exif_data['lens_model'] = str(tags.get('EXIF LensModel', None))
            exif_data['image_description'] = str(tags.get('Image ImageDescription', None))
    except Exception as e:
        print(f"Error extracting EXIF data from {filepath}: {e}")
    return exif_data

def convert_to_degrees(value) -> float:
    d = value.values[0].num / value.values[0].den
    m = value.values[1].num / value.values[1].den
    s = value.values[2].num / value.values[2].den
    return d + (m / 60.0) + (s / 3600.0)

def insert_exif_data(filepath: str, directory: str, exif_data: Dict[str, Optional[Any]]) -> int:
    cur.execute("""
    INSERT INTO exif_metadata (filepath, directory, datetime, datetime_original, gps_latitude, gps_latitude_ref, gps_longitude, gps_longitude_ref, make, model, lens_model, image_description)
    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) RETURNING id
    """, (
        filepath.encode('utf-8').decode('utf-8'), 
        directory.encode('utf-8').decode('utf-8'),
        exif_data.get('datetime'),
        exif_data.get('datetime_original'),
        exif_data.get('gps_latitude'),
        exif_data.get('gps_latitude_ref'),
        exif_data.get('gps_longitude'),
        exif_data.get('gps_longitude_ref'),
        exif_data.get('make').encode('utf-8').decode('utf-8') if exif_data.get('make') else None,
        exif_data.get('model').encode('utf-8').decode('utf-8') if exif_data.get('model') else None,
        exif_data.get('lens_model').encode('utf-8').decode('utf-8') if exif_data.get('lens_model') else None,
        exif_data.get('image_description').encode('utf-8').decode('utf-8') if exif_data.get('image_description') else None
    ))
    image_id: int = cur.fetchone()[0]
    conn.commit()
    return image_id

def create_keywords_index(image_id: int, text: str) -> None:
    keywords = re.split(r'[\W\-_;,.]+', text.lower())
    for keyword in set(keywords):
        if keyword:  # Ensure no empty keywords
            cur.execute("INSERT INTO keywords (keyword, image_id) VALUES (%s, %s)", (keyword, image_id))
    conn.commit()

def process_directory(directory: str) -> None:
    for root, _, files in os.walk(directory):
        print(root)
        for file in files:
            filepath = os.path.join(root, file)
            if file.lower().endswith(('.jpg', '.jpeg', '.png', '.tiff', '.bmp', '.gif')):
                exif_data = extract_exif_data(filepath)
                if exif_data:
                    dir_name = os.path.basename(root)
                    image_id = insert_exif_data(filepath, dir_name, exif_data)
                    if exif_data.get('image_description'):
                        create_keywords_index(image_id, exif_data['image_description'])
                    create_keywords_index(image_id, dir_name)

DIRECTORY_PATH = '/Volumes/Photos/exports'
process_directory(DIRECTORY_PATH)

# Close the database connection
cur.close()
conn.close()
