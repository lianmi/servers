<template>
<div>
    <el-form :model="formData" label-position="right" label-width="80px">
             <el-form-item label="允许撤单:">
                <el-switch active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" v-model="formData.allowCancel" clearable ></el-switch>
          </el-form-item>
           
             <el-form-item label="详情图片1:">
                <el-input v-model="formData.descPic1" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="详情图片2:">
                <el-input v-model="formData.descPic2" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="详情图片3:">
                <el-input v-model="formData.descPic3" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="详情图片4:">
                <el-input v-model="formData.descPic4" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="详情图片5:">
                <el-input v-model="formData.descPic5" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="descPic6字段:">
                <el-input v-model="formData.descPic6" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="修改时间:"><el-input v-model.number="formData.modifyAt" clearable placeholder="请输入"></el-input>
          </el-form-item>
           
             <el-form-item label="商品介绍:">
                <el-input v-model="formData.productDesc" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="商品UUID:">
                <el-input v-model="formData.productId" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="商品名称:">
                <el-input v-model="formData.productName" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="商品大图1:">
                <el-input v-model="formData.productPic1Large" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="商品大图2:">
                <el-input v-model="formData.productPic2Large" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="商品大图3:">
                <el-input v-model="formData.productPic3Large" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="商品类型:">
                 <el-select v-model="formData.productType" placeholder="请选择" clearable>
                     <el-option v-for="(item,key) in intOptions" :key="key" :label="item.label" :value="item.value"></el-option>
                 </el-select>
          </el-form-item>
           
             <el-form-item label="shortVideo字段:">
                <el-input v-model="formData.shortVideo" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           <el-form-item>
           <el-button @click="save" type="primary">保存</el-button>
           <el-button @click="back" type="primary">返回</el-button>
           </el-form-item>
    </el-form>
</div>
</template>

<script>
import {
    createGeneralProduct,
    updateGeneralProduct,
    findGeneralProduct
} from "@/api/generalProducts";  //  此处请自行替换地址
import infoList from "@/mixins/infoList";
export default {
  name: "GeneralProduct",
  mixins: [infoList],
  data() {
    return {
      type: "",
      intOptions:[],
          formData: {
            allowCancel:false,
            descPic1:"",
            descPic2:"",
            descPic3:"",
            descPic4:"",
            descPic5:"",
            descPic6:"",
            modifyAt:0,
            productDesc:"",
            productId:"",
            productName:"",
            productPic1Large:"",
            productPic2Large:"",
            productPic3Large:"",
            productType:0,
            shortVideo:"",
            
      }
    };
  },
  methods: {
    async save() {
      let res;
      switch (this.type) {
        case "create":
          res = await createGeneralProduct(this.formData);
          break;
        case "update":
          res = await updateGeneralProduct(this.formData);
          break;
        default:
          res = await createGeneralProduct(this.formData);
          break;
      }
      if (res.code == 0) {
        this.$message({
          type:"success",
          message:"创建/更改成功"
        })
      }
    },
    back(){
        this.$router.go(-1)
    }
  },
  async created() {
   // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if(this.$route.query.id){
    const res = await findGeneralProduct({ ID: this.$route.query.id })
    if(res.code == 0){
       this.formData = res.data.regeneralProducts
       this.type == "update"
     }
    }else{
     this.type == "create"
   }
  
    await this.getDict("int");
    
}
};
</script>

<style>
</style>