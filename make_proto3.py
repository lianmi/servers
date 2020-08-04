import os
# 搜索路径
CONFIG_LIST = {
    # 搜索的 根目录 , 相对路径或者绝对路径
    "protoroot":"protofile/proto",
    # 生成的路径
    "output": "/home/wujehy/PycharmProjects/testoutput",
    # 根目录下需要生成的子模块
    "modules":[
        "auth",
        "avchat",
        "code",
        "friends",
        "msg",
        "signal",
        "syn",
        "system",
        "team",
        "tribe",
        "user"
    ]
}
# 生成的类型 参数
CONFOG_GENERATE_TYPES = [
    # java
    {
        "flag": "--java_out",
        "output":"jout"
    },
    # objc
    {
        "flag": "--objc_out",
        "output": "ios"
    },
    # cpp
    {
        "flag": "--cpp_out",
        "output": "cpp"
    },

]

def mkdir(path):
    # 去除首位空格
    path = path.strip()
    # 去除尾部 \ 符号
    path = path.rstrip("\\")

    # 判断路径是否存在
    # 存在     True
    # 不存在   False
    isExists = os.path.exists(path)

    # 判断结果
    if not isExists:
        # 如果不存在则创建目录
        # 创建目录操作函数
        os.makedirs(path)

        print(path + ' 创建成功')
        return True
    else:
        # 如果目录存在则不创建，并提示目录已存在
        print(path + ' 目录已存在')
        return False


# 查找目录下的文件
def find_files(path):

    print(os.listdir(path))

# 生成 protobuf
def generate_protobuf(filepath , protoc ="protoc" , generate = True ):
    flags =""
    for config in CONFOG_GENERATE_TYPES :
        flags += " %s=%s/%s" %(config.get("flag") ,CONFIG_LIST["output"]  , config.get("output"))
    command = "%s --proto_path=%s %s %s " %(protoc ,CONFIG_LIST["protoroot"] , flags , filepath)
    print("command : " , command)

    if generate :
        print("doing  ... ")
        os.system(command)
        print("end ")






def do_run():
    print("start generate ... ")
    # go to output dir
    mkdir(CONFIG_LIST["output"])
    # 生成各种类型的输出路径
    for config in CONFOG_GENERATE_TYPES:
        output_path = "%s/%s" %(CONFIG_LIST["output"] , config.get("output"))
        print("mkdir : " , output_path)
        mkdir(output_path)
    # cd_command = "cd %s" % (CONFIG_LIST["output"])
    # os.system(cd_command)
    for mod in CONFIG_LIST["modules"]:
        print("current module : "  , mod )
        current_dir = "%s/%s" %(CONFIG_LIST["protoroot"] , mod)
        file_lists = os.listdir(current_dir)
        # print("current files "  ,  file_lists)
        for file in file_lists :
            relative_path = "%s/%s" % (mod , file)
            generate_protobuf(relative_path)


if __name__ == "__main__":

    do_run()

