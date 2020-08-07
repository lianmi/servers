import os
CONFIG_LIST = {
    "protoroot":".",
    "output": "../im-mqtt-sdk/src",
    "modules":[
        "auth",
        # "code",
        "friends",
        "global",
        "msg",
        # "signal",
        "syn",
        "team",
        "user"
    ]
}


CONFOG_GENERATE_TYPES = [
    # cpp
    {
        "flag": "--cpp_out",
        "output": ""
    },

]

def mkdir(path):
    path = path.strip()
    path = path.rstrip("\\")
    isExists = os.path.exists(path)

    if not isExists:
        os.makedirs(path)

        print(path + 'creat success')
        return True
    else:
        print(path + 'path existence \n')
        return True


def find_files(path):

    print(os.listdir(path))

def generate_protobuf(filepath , protoc ="protoc" , generate = True ):
    flags =""
    for config in CONFOG_GENERATE_TYPES :
        flags += " %s=%s/%s" %(config.get("flag"), CONFIG_LIST["output"]  , config.get("output"))
    command = "%s --proto_path=%s %s %s" %(protoc, CONFIG_LIST["protoroot"] , flags , filepath)
    print("command : " , command)

    if generate :
        # print("doing  ... ")
        os.system(command)
        # print("end ")


def do_run():
    print("start generate ... ")
    # go to output dir
    # mkdir(CONFIG_LIST["output"])
    for config in CONFOG_GENERATE_TYPES:
        output_path = "%s/%s" %(CONFIG_LIST["output"] , config.get("output"))
        print("mkdir : " , output_path)
        mkdir(output_path)
    # cd_command = "cd %s" % (CONFIG_LIST["output"])
    # os.system(cd_command)
    for mod in CONFIG_LIST["modules"]:
        print("current module : "  , mod )
        current_dir = "api/proto/%s" %(mod)
        file_lists = os.listdir(current_dir)
        # print("current files "  ,  file_lists)
        for file in file_lists :
            if os.path.splitext(file)[1] == '.proto': 
                relative_path = "api/proto/%s/%s" % (mod , file)
                generate_protobuf(relative_path)


if __name__ == "__main__":

    do_run()

