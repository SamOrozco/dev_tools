import os

binaryLocation = "../binaries/"
goBuildLocation = ".."
goOsGoArch = [
    ('linux', '386'),
    ('linux', 'amd64'),
    ('linux', 'arm'),
    ('linux', 'arm64'),
    ('windows', '386'),
    ('windows', 'amd64'),
    ('windows', 'arm'),
    ('darwin', 'amd64'),
    ('darwin', 'arm'),
    ('darwin', 'arm64'),
]


def buildCombinationsOfGoBinaries(combinations, buildLocaton, binaryLocation):
    # build binary current running machine
    goBuildBinaryWithNameAndDir(binaryLocation + getBinaryName(os.name == 'nt'), buildLocaton)
    for goos, arch in combinations:
        os.environ['GOOS'] = goos
        os.environ['GOARCH'] = arch
        binaryName = binaryLocation + goos + '_' + arch + '_' + getBinaryName(goos == 'windows')
        print(f'building...{binaryName}')
        goBuildBinaryWithNameAndDir(binaryName, buildLocaton)
        print('success')


def goBuildBinaryWithNameAndDir(name, dir):
    stream = os.popen(f"go build -o {name} {dir}")


def getBinaryName(windows):
    if windows:
        return 'ew.exe'
    else:
        return 'ew'


# Press the green button in the gutter to run the script.
if __name__ == '__main__':
    buildCombinationsOfGoBinaries(goOsGoArch, "..", "../binaries/")

# See PyCharm help at https://www.jetbrains.com/help/pycharm/
