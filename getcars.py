import wget
import os
import sys
import requests
import concurrent.futures
import chompjs
import re

cwd = os.getcwd()
base = "https://awesomecars.neocities.org"

ILLEGAL_RE = re.compile(r'[/\?<>\\:\*\|":]')
CONTROL_RE = re.compile(r'[\x00-\x1f\x80-\x9f]')
RESERVED_RE = re.compile(r'^\.+$')
WINDOWS_RESERVED_RE = re.compile(
    r'(?i)^(con|prn|aux|nul|com[0-9]|lpt[0-9])(\..*)?$')
WINDOWS_TRAILING_RE = re.compile(r'^\.+$')


def filename_sanitize(filename: str) -> str:
    to_return = re.sub(ILLEGAL_RE, "", filename)
    to_return = re.sub(CONTROL_RE, "", to_return)
    to_return = re.sub(RESERVED_RE, "", to_return)
    if os.name == 'nt':
        to_return = re.sub(WINDOWS_RESERVED_RE, "", to_return)
        to_return = re.sub(WINDOWS_TRAILING_RE, "", to_return)
    return to_return


def download_cars(car: str) -> None:
    download_dir = cwd + '/cars_v3'
    split_name = car.split(' - ')
    num = split_name[0].replace('#','').replace(',', '')
    name = split_name[1] if len(split_name) == 2 else ' - '.join(split_name[1:])
    new_filename = f'{download_dir}/{num.zfill(4)}-{filename_sanitize(name.replace(" ","_"))}.mp4'
    if os.path.exists(new_filename):
        return
    wget.download(f'{base}/ver2/{num}.mp4', new_filename)
    print(f'\r\ndownloaded car {num}: {name}')


def main() -> None:
    with concurrent.futures.ProcessPoolExecutor() as executor:
        if not os.path.isdir(cwd + "/cars_v3"):
            print("creating cars_v3 directory")
            os.mkdir(cwd + "/cars_v3")
        js = requests.get(f'{base}/search.js').text
        objects = chompjs.parse_js_objects(js)
        cars: list[str] = next(objects)
        executor.map(download_cars, cars, chunksize=16)


if __name__ == '__main__':
    try:
        main()
    except KeyboardInterrupt:
        print("\ninterrupted. cleaning temp files...")
        try:
            for filename in os.listdir(cwd + '/cars_v3'):
                f = os.path.join(cwd, filename)
                if os.path.isfile(f):
                    if f.endswith(".tmp"):
                        os.remove(f)
        except SystemExit:
            sys.exit(0)
