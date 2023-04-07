import subprocess
import datetime

format = '%Y/%m/%d %H:%M:%S.%f'

PACK = "./out/pack"  # location of pack binary

username = "edithwu"

N = 5
k = 2

labels = {"build": "build",
          "lifecycle": "lifecycleExecutor",
          "pullImage": "NetWork I/O for pulling image",
          "download": "NetWork I/O for downloading",
          "saveBuilder": "I/O for saving builder",
          "parseImageReference": "parse image reference",
          "processAppPath": "process app path",
          "processProxyConfig": "process proxy config",
          "processBuilderName": "process builder name",
          "getBuilder": "get builder",
          "resolveRunImage": "resolve run image",
          "validateRunImage": "validate run image",
          "processBuildpack": "process buildpack",
          "validateMixins": "validate mixins",
          "createEphemeralBuilder": "create ephemeral builder",
          "processVolume": "process volume",
          "getFileFilter": "get file filter",
          "translateRegistry": "translate registry",
          "stackMixins": "stack mixins",
          "allBuildpacks": "all buildpacks",
          "assembleAvailableMixins": "assemble availble mixins",
          "ensureStackSupport": "ensure stack support",
          "createNewBuilder": "create new builder",
          "addBuildpack": "add buildpack",
          "outputBuildpack": "output buildpack",
          "makeDirectoryTemp": "make directory temp",
          "getDefaultDirectoryLayer": "get default directory layer",
          "addDefaultDirectoryLayer": "add default directory layer",
          "validateBuildpacks": "validate buildpacks",
          "validateExtensions": "validate extensions",
          "getBuildpackLabel": "get buildpack label",
          "addBuildpackModule": "add buildpack module",
          "setBuildpackLabel": "set buildpack label",
          "getExtensionLabel": "get extension label",
          "addExtensionModule": "add extension module",
          "setExtensionLabel": "set extension label",
          "getStackLayer": "get stack layer",
          "addStackLayer": "add stack layer",
          "getEnvironmentLayer": "get environment layer",
          "addEnvironmentLayer": "add environment layer",
          "setMetaLabel": "set meta label",
          "setMixinsLabel": "set mixins label",
          "setWorkingDirectory": "set working directory",
          "getLifecycleLayer": "get lifecycle layer",
          "addLifecycleLayer": "add lifecycle layer",
          "processBuildpackOrder": "process buildpack order",
          "processExtensionOrder": "process extension order",
          "getOrderLayer": "get order layer",
          "addOrderLayer": "add order layer",
          "setBuildpackOrderLayer": "set buildpack order layer",
          "setExtensionOrderLayer": "set extension order layer"
          }

def run(command, out, error="error.out"):
    print(command)
    process = subprocess.Popen(command, shell=True, stdout=open(out, "w"), stderr=open(error, "w"))

    process.wait()

    output = open(out).read()
    error = open(error).read()

    if error != "":
        print(error)

    def calculateTime(label):
        time = datetime.timedelta()
        start = output.find(label + " start")
        while start != -1:
            startTime = datetime.datetime.strptime(output[output.rfind('\n', 0, start):start].strip(), format)
            end = output.find(label + " end", start)
            endTime = datetime.datetime.strptime(output[output.rfind('\n', 0, end):end].strip(), format)
            time += endTime - startTime

            # remove time spent on pulling image from validating run image and process buildpack
            if label == labels["validateRunImage"] or label == labels["processBuildpack"]:
                start = output.find(labels["pullImage"] + " start", start, end)
                if start != -1:
                    startTime = datetime.datetime.strptime(output[output.rfind('\n', 0, start):start].strip(), format)
                    end = output.find(labels["pullImage"] + " end", start)
                    endTime = datetime.datetime.strptime(output[output.rfind('\n', 0, end):end].strip(), format)
                    time -= endTime - startTime

            # remove time spent on saving builder from creating ephemeral builder
            if label == labels["createEphemeralBuilder"]:
                start = output.find(labels["saveBuilder"] + " start", start, end)
                if start != -1:
                    startTime = datetime.datetime.strptime(output[output.rfind('\n', 0, start):start].strip(), format)
                    end = output.find(labels["saveBuilder"] + " end", start)
                    endTime = datetime.datetime.strptime(output[output.rfind('\n', 0, end):end].strip(), format)
                    time -= endTime - startTime

            start = output.find(label + " start", end)

        return time

    def getBuildpackNumber():
        numBuildpack = "0"
        start = output.find("the number of buildpacks: ")
        if start != -1:
            numBuildpack = output[output.find(":", start) + 2:output.find("\n", start)]
        return numBuildpack

    result = {}
    for key in labels.keys():
        result[key] = calculateTime(labels[key])

    result["BuildpackNum"] = getBuildpackNumber()

    return result


def repeat(command, out):
    repeatResult = {}
    for key in labels.keys():
        repeatResult[key] = datetime.timedelta(0)
    for i in range(N):
        result = run(command, out)
        if i < k:
            continue
        else:
            for key in repeatResult.keys():
                repeatResult[key] += result[key]

    for value in repeatResult.values():
        value /= N - k
    return repeatResult


def firstBuild(imageName):
    builder = "paketobuildpacks/builder:base"
    repeatResult = {}
    for key in labels.keys():
        repeatResult[key] = datetime.timedelta(0)
    for i in range(N):
        command = PACK + " build " + imageName + "-" + datetime.datetime.now().strftime("%S.%f") + \
                  " --builder " + builder + " --timestamps -v"
        result = run(command, "first_build.out")
        if i < k:
            continue
        else:
            for key in repeatResult.keys():
                repeatResult[key] += result[key]

    for value in repeatResult.values():
        value /= N - k
    return repeatResult


def laterBuild(imageName):
    builder = "paketobuildpacks/builder:base"
    command = PACK + " build " + imageName + " --builder " + builder + " --timestamps -v"
    return repeat(command, "later_build.out")


def tinyBuild(imageName):
    builder = "paketobuildpacks/builder:tiny"
    command = PACK + " build " + imageName + " --builder " + builder + " --timestamps -v"
    return repeat(command, "tiny_build.out")


def buildpackBuild(imageName):
    buildpack = "docker://cnbs/sample-package:hello-universe"
    builder = "paketobuildpacks/builder:tiny"
    command = PACK + " build " + imageName + " --builder " + builder + " --buildpack " + buildpack + " --timestamps -v"
    return repeat(command, "buildpack_build.out")


def cacheImageBuild(imageName):
    builder = "paketobuildpacks/builder:base"
    command = PACK + " build " + imageName + " --builder " + builder + \
              " --timestamps -v --cache type=build;format=image;name=paketo-demo-app;"
    return repeat(command, "cache_image_build.out")


def neverBuild(imageName):
    builder = "paketobuildpacks/builder:base"
    command = PACK + " build " + imageName + " --builder " + builder + " --pull-policy never --timestamps -v"
    return repeat(command, "never_policy_build.out")


def alwaysBuild(imageName):
    builder = "paketobuildpacks/builder:base"
    command = PACK + " build " + imageName + " --builder " + builder + " --pull-policy always --timestamps -v"
    return repeat(command, "always_policy_build.out")


def publishBuild(imageName):
    builder = "paketobuildpacks/builder:base"
    command = PACK + " build docker.io/" + username + "/" + imageName + ":latest --builder " + builder + " --publish --timestamps -v"
    return repeat(command, "publish_build.out")


def untrustedBuild(imageName):
    origin = "paketobuildpacks/builder:base"
    builder = "mybuilder:base"
    command = "docker tag " + origin + " " + builder + " && " + \
              PACK + " build " + imageName + " --builder " + builder + " --timestamps -v"
    return repeat(command, "untrusted_build.out")


def profilingTime():
    file = open("profiling.csv", "w")
    file.write("condition")
    for key in labels.keys():
        file.write(", " + key)
    file.write("\n")

    imageName = "paketo-demo-app"

    def output(taskName, result):
        file.write(taskName)
        for key in labels.keys():
            file.write(", " + str(result[key]))
        file.write("\n")

    result = firstBuild(imageName)
    output("first build", result)

    result = laterBuild(imageName)
    output("later build", result)

    result = tinyBuild(imageName)
    output("tiny build", result)

    result = buildpackBuild(imageName)
    output("buildpack build", result)

    result = cacheImageBuild(imageName)
    output("cache image build", result)

    result = neverBuild(imageName)
    output("never policy build", result)

    result = alwaysBuild(imageName)
    output("always policy build", result)

    result = publishBuild(imageName)
    output("publish build", result)

    result = untrustedBuild(imageName)
    output("untrusted build", result)

    file.close()

    return


def differentBuilder():
    imageName = "paketo-demo-app"
    paketoBuilder = "paketobuildpacks/builder"
    paketoTags = ["tiny", "base", "full"]
    cnfBuilder = "cnbs/sample-builder"
    cnfTags = ["wine", "jammy", "alpine", "bionic"]

    outputTime = {}
    buildpackNum = {}

    def calculateTimeAndNum(tags, builder):
        for tag in tags:
            command = PACK + " build " + imageName + " --builder " + builder + ":" + tag + " --timestamps -v"
            outputBuildpackTime = datetime.timedelta(0)
            for i in range(N):
                result = run(command, "differentbuilder.out")
                if i < k:
                    buildpackNum[tag] = result["BuildpackNum"]
                else:
                    outputBuildpackTime += result["outputBuildpack"]

            outputBuildpackTime /= N - k
            outputTime[tag] = outputBuildpackTime

    calculateTimeAndNum(paketoTags, paketoBuilder)
    calculateTimeAndNum(cnfTags, cnfBuilder)

    process = subprocess.Popen("docker image ls", shell=True, stdout=open("dockerImage.out", "w"), stderr=open("error.out", "w"))
    process.wait()
    output = open("dockerImage.out").read()
    error = open("error.out").read()

    if error != "":
        print(error)

    for tag in paketoTags:
        start = output.find("paketobuildpacks/builder   " + tag)
        end = output.find("\n", start)
        size = output[output.rfind(" ", start, end) + 1: end]
        print(paketoBuilder + ":" + tag + ", " + size + ", " + buildpackNum[tag] + ", " + str(outputTime[tag]))

    for tag in cnfTags:
        start = output.find("cnbs/sample-builder        " + tag)
        end = output.find("\n", start)
        size = output[output.rfind(" ", start, end) + 1: end]
        print(cnfBuilder + ":" + tag + ", " + size + ", " + buildpackNum[tag] + ", " + str(outputTime[tag]))

    return

def main():

    profilingTime()

    # differentBuilder()


if __name__ == "__main__":
    main()
