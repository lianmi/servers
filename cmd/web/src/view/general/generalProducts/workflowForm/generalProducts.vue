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
           <el-button v-if="this.wf.clazz == 'start'" @click="start" type="primary">启动</el-button>
           <!-- complete传入流转参数 决定下一步会流转到什么位置 此处可以设置多个按钮来做不同的流转 -->
           <el-button v-if="canShow" @click="complete('yes')" type="primary">提交</el-button>
           <el-button v-if="showSelfNode" @click="complete('')" type="primary">确认</el-button>
           <el-button @click="back" type="primary">返回</el-button>
           </el-form-item>
    </el-form>
</div>
</template>

<script>
import {
    startWorkflow,
    completeWorkflowMove
} from "@/api/workflowProcess";
import infoList from "@/mixins/infoList";
import { mapGetters } from "vuex";
export default {
  name: "GeneralProduct",
  mixins: [infoList],
  props:{
      business:{
         type:Object,
        default:function(){return null}
      },
      wf:{
        type:Object,
        default:function(){return{}}
      },
      move:{
         type:Object,
         default:function(){return{}}
      },
      workflowMoveID:{
        type:[Number,String],
        default:0
      }
   },
  data() {
    return {
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
  computed:{
      showSelfNode(){
         if(this.wf.assignType == "self" && this.move.promoterID == this.userInfo.ID){
             return true
         }else{
             return false
         }
      },
      canShow(){
         if(this.wf.assignType == "user"){
            if(this.wf.assignValue.indexOf(","+this.userInfo.ID+",")>-1 && this.wf.clazz == 'userTask'){
               return true
            }else{
               return false
            }
         }else if(this.wf.assign_type == "authority"){
            if(this.wf.assignValue.indexOf(","+this.userInfo.authorityId+",")>-1 && this.wf.clazz == 'userTask'){
               return true
            }else{
               return false
            }
         }
      },
      ...mapGetters("user", ["userInfo"])
  },
  methods: {
    async start() {
      const res = await startWorkflow({
            business:this.formData,
            wf:{
              workflowMoveID:this.workflowMoveID,
              businessId:0,
              businessType:"generalProducts",
              workflowProcessID:this.wf.workflowProcessID,
              workflowNodeID:this.wf.id,
              promoterID:this.userInfo.ID,
              operatorID:this.userInfo.ID,
              action:"create",
              param:""
              }
          });
      if (res.code == 0) {
        this.$message({
          type:"success",
          message:"启动成功"
        })
       this.back()
      }
    },
    async complete(param){
     const res = await completeWorkflowMove({
            business:this.formData,
            wf:{
              workflowMoveID:this.workflowMoveID,
              businessID:this.formData.ID,
              businessType:"generalProducts",
              workflowProcessID:this.wf.workflowProcessID,
              workflowNodeID:this.wf.id,
              promoterID:this.userInfo.ID,
              operatorID:this.userInfo.ID,
              action:"complete",
              param:param
              }
     })
     if(res.code == 0){
       this.$message({
          type:"success",
          message:"提交成功"
       })
       this.back()
     }
    },
    back(){
        this.$router.go(-1)
    }
  },
  async created() {
    if(this.business){
     this.formData = this.business
    }
}
};
</script>

<style>
</style>