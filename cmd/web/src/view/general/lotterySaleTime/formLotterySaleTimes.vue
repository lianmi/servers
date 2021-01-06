<template>
<div>
    <el-form :model="formData" label-position="right" label-width="80px">
             <el-form-item label="类型:"><el-input v-model.number="formData.lotteryType" clearable placeholder="请输入"></el-input>
          </el-form-item>
           
             <el-form-item label="彩票名称:">
                <el-input v-model="formData.lotteryName" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="开始的星期数:">
                <el-input v-model="formData.saleStartWeek" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="开售时刻:">
                <el-input v-model="formData.saleStartTime" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="停售星期数:">
                <el-input v-model="formData.saleEndWeek" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="停售时刻 :">
                <el-input v-model="formData.saleEndTime" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="节假日停售(选填):">
                <el-input v-model="formData.holidays" clearable placeholder="请输入" ></el-input>
          </el-form-item>
           
             <el-form-item label="是否激活:">
                <el-switch active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" v-model="formData.isActive" clearable ></el-switch>
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
    createLotterySaleTimes,
    updateLotterySaleTimes,
    findLotterySaleTimes
} from "@/api/lotterySaleTimes";  //  此处请自行替换地址
import infoList from "@/mixins/infoList";
export default {
  name: "LotterySaleTimes",
  mixins: [infoList],
  data() {
    return {
      type: "",formData: {
            lotteryType:0,
            lotteryName:"",
            saleStartWeek:"",
            saleStartTime:"",
            saleEndWeek:"",
            saleEndTime:"",
            holidays:"",
            isActive:false,
            
      }
    };
  },
  methods: {
    async save() {
      let res;
      switch (this.type) {
        case "create":
          res = await createLotterySaleTimes(this.formData);
          break;
        case "update":
          res = await updateLotterySaleTimes(this.formData);
          break;
        default:
          res = await createLotterySaleTimes(this.formData);
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
    const res = await findLotterySaleTimes({ ID: this.$route.query.id })
    if(res.code == 0){
       this.formData = res.data.relotterySaleTimes
       this.type == "update"
     }
    }else{
     this.type == "create"
   }
  
}
};
</script>

<style>
</style>