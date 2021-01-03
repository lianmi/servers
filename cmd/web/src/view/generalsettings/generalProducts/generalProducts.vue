<template>
  <div>
    <div class="search-term">
      <el-form :inline="true" :model="searchInfo" class="demo-form-inline">                                                
        <el-form-item>
          <el-button @click="onSubmit" type="primary">查询</el-button>
        </el-form-item>
        <el-form-item>
          <el-button @click="openDialog" type="primary">新增generalProducts表</el-button>
        </el-form-item>
        <el-form-item>
          <el-popover placement="top" v-model="deleteVisible" width="160">
            <p>确定要删除吗？</p>
              <div style="text-align: right; margin: 0">
                <el-button @click="deleteVisible = false" size="mini" type="text">取消</el-button>
                <el-button @click="onDelete" size="mini" type="primary">确定</el-button>
              </div>
            <el-button icon="el-icon-delete" size="mini" slot="reference" type="danger">批量删除</el-button>
          </el-popover>
        </el-form-item>
      </el-form>
    </div>
    <el-table
      :data="tableData"
      @selection-change="handleSelectionChange"
      border
      ref="multipleTable"
      stripe
      style="width: 100%"
      tooltip-effect="dark"
    >
    <el-table-column type="selection" width="55"></el-table-column>
    <el-table-column label="日期" width="180">
         <template slot-scope="scope">{{scope.row.CreatedAt|formatDate}}</template>
    </el-table-column>
    
    <el-table-column label="allowCancel字段" prop="allowCancel" width="120">
         <template slot-scope="scope">{{scope.row.allowCancel|formatBoolean}}</template>
    </el-table-column>
    
    <el-table-column label="createAt字段" prop="createAt" width="120"></el-table-column> 
    
    <el-table-column label="descPic1字段" prop="descPic1" width="120"></el-table-column> 
    
    <el-table-column label="descPic2字段" prop="descPic2" width="120"></el-table-column> 
    
    <el-table-column label="descPic3字段" prop="descPic3" width="120"></el-table-column> 
    
    <el-table-column label="descPic4字段" prop="descPic4" width="120"></el-table-column> 
    
    <el-table-column label="descPic5字段" prop="descPic5" width="120"></el-table-column> 
    
    <el-table-column label="descPic6字段" prop="descPic6" width="120"></el-table-column> 
    
    <el-table-column label="modifyAt字段" prop="modifyAt" width="120"></el-table-column> 
    
    <el-table-column label="productDesc字段" prop="productDesc" width="120"></el-table-column> 
    
    <el-table-column label="productId字段" prop="productId" width="120"></el-table-column> 
    
    <el-table-column label="productName字段" prop="productName" width="120"></el-table-column> 
    
    <el-table-column label="productPic1Large字段" prop="productPic1Large" width="120"></el-table-column> 
    
    <el-table-column label="productPic1Middle字段" prop="productPic1Middle" width="120"></el-table-column> 
    
    <el-table-column label="productPic1Small字段" prop="productPic1Small" width="120"></el-table-column> 
    
    <el-table-column label="productPic2Large字段" prop="productPic2Large" width="120"></el-table-column> 
    
    <el-table-column label="productPic2Middle字段" prop="productPic2Middle" width="120"></el-table-column> 
    
    <el-table-column label="productPic2Small字段" prop="productPic2Small" width="120"></el-table-column> 
    
    <el-table-column label="productPic3Large字段" prop="productPic3Large" width="120"></el-table-column> 
    
    <el-table-column label="productPic3Middle字段" prop="productPic3Middle" width="120"></el-table-column> 
    
    <el-table-column label="productPic3Small字段" prop="productPic3Small" width="120"></el-table-column> 
    
    <el-table-column label="productType字段" prop="productType" width="120"></el-table-column> 
    
    <el-table-column label="shortVideo字段" prop="shortVideo" width="120"></el-table-column> 
    
    <el-table-column label="thumbnail字段" prop="thumbnail" width="120"></el-table-column> 
    
      <el-table-column label="按钮组">
        <template slot-scope="scope">
          <el-button class="table-button" @click="updateGeneralProduct(scope.row)" size="small" type="primary" icon="el-icon-edit">变更</el-button>
          <el-popover placement="top" width="160" v-model="scope.row.visible">
            <p>确定要删除吗？</p>
            <div style="text-align: right; margin: 0">
              <el-button size="mini" type="text" @click="scope.row.visible = false">取消</el-button>
              <el-button type="primary" size="mini" @click="deleteGeneralProduct(scope.row)">确定</el-button>
            </div>
            <el-button type="danger" icon="el-icon-delete" size="mini" slot="reference">删除</el-button>
          </el-popover>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      :current-page="page"
      :page-size="pageSize"
      :page-sizes="[10, 30, 50, 100]"
      :style="{float:'right',padding:'20px'}"
      :total="total"
      @current-change="handleCurrentChange"
      @size-change="handleSizeChange"
      layout="total, sizes, prev, pager, next, jumper"
    ></el-pagination>

    <el-dialog :before-close="closeDialog" :visible.sync="dialogFormVisible" title="弹窗操作">
      <el-form :model="formData" label-position="right" label-width="80px">
         <el-form-item label="allowCancel字段:">
            <el-switch active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" v-model="formData.allowCancel" clearable ></el-switch>
      </el-form-item>
       
         <el-form-item label="createAt字段:"><el-input v-model.number="formData.createAt" clearable placeholder="请输入"></el-input>
      </el-form-item>
       
         <el-form-item label="descPic1字段:">
            <el-input v-model="formData.descPic1" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="descPic2字段:">
            <el-input v-model="formData.descPic2" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="descPic3字段:">
            <el-input v-model="formData.descPic3" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="descPic4字段:">
            <el-input v-model="formData.descPic4" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="descPic5字段:">
            <el-input v-model="formData.descPic5" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="descPic6字段:">
            <el-input v-model="formData.descPic6" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="modifyAt字段:"><el-input v-model.number="formData.modifyAt" clearable placeholder="请输入"></el-input>
      </el-form-item>
       
         <el-form-item label="productDesc字段:">
            <el-input v-model="formData.productDesc" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productId字段:">
            <el-input v-model="formData.productId" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productName字段:">
            <el-input v-model="formData.productName" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic1Large字段:">
            <el-input v-model="formData.productPic1Large" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic1Middle字段:">
            <el-input v-model="formData.productPic1Middle" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic1Small字段:">
            <el-input v-model="formData.productPic1Small" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic2Large字段:">
            <el-input v-model="formData.productPic2Large" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic2Middle字段:">
            <el-input v-model="formData.productPic2Middle" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic2Small字段:">
            <el-input v-model="formData.productPic2Small" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic3Large字段:">
            <el-input v-model="formData.productPic3Large" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic3Middle字段:">
            <el-input v-model="formData.productPic3Middle" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productPic3Small字段:">
            <el-input v-model="formData.productPic3Small" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="productType字段:"><el-input v-model.number="formData.productType" clearable placeholder="请输入"></el-input>
      </el-form-item>
       
         <el-form-item label="shortVideo字段:">
            <el-input v-model="formData.shortVideo" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       
         <el-form-item label="thumbnail字段:">
            <el-input v-model="formData.thumbnail" clearable placeholder="请输入" ></el-input>
      </el-form-item>
       </el-form>
      <div class="dialog-footer" slot="footer">
        <el-button @click="closeDialog">取 消</el-button>
        <el-button @click="enterDialog" type="primary">确 定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import {
    createGeneralProduct,
    deleteGeneralProduct,
    deleteGeneralProductByIds,
    updateGeneralProduct,
    findGeneralProduct,
    getGeneralProductList
} from "@/api/generalProducts";  //  此处请自行替换地址
import { formatTimeToStr } from "@/utils/date";
import infoList from "@/mixins/infoList";
export default {
  name: "GeneralProduct",
  mixins: [infoList],
  data() {
    return {
      listApi: getGeneralProductList,
      dialogFormVisible: false,
      visible: false,
      type: "",
      deleteVisible: false,
      multipleSelection: [],formData: {
            allowCancel:false,
            createAt:0,
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
            productPic1Middle:"",
            productPic1Small:"",
            productPic2Large:"",
            productPic2Middle:"",
            productPic2Small:"",
            productPic3Large:"",
            productPic3Middle:"",
            productPic3Small:"",
            productType:0,
            shortVideo:"",
            thumbnail:"",
            
      }
    };
  },
  filters: {
    formatDate: function(time) {
      if (time != null && time != "") {
        var date = new Date(time);
        return formatTimeToStr(date, "yyyy-MM-dd hh:mm:ss");
      } else {
        return "";
      }
    },
    formatBoolean: function(bool) {
      if (bool != null) {
        return bool ? "是" :"否";
      } else {
        return "";
      }
    }
  },
  methods: {
      //条件搜索前端看此方法
      onSubmit() {
        this.page = 1
        this.pageSize = 10      
        if (this.searchInfo.allowCancel==""){
          this.searchInfo.allowCancel=null
        }                             
        this.getTableData()
      },
      handleSelectionChange(val) {
        this.multipleSelection = val
      },
      async onDelete() {
        const ids = []
        if(this.multipleSelection.length == 0){
          this.$message({
            type: 'warning',
            message: '请选择要删除的数据'
          })
          return
        }
        this.multipleSelection &&
          this.multipleSelection.map(item => {
            ids.push(item.ID)
          })
        const res = await deleteGeneralProductByIds({ ids })
        if (res.code == 0) {
          this.$message({
            type: 'success',
            message: '删除成功'
          })
          if (this.tableData.length == ids.length) {
              this.page--;
          }
          this.deleteVisible = false
          this.getTableData()
        }
      },
    async updateGeneralProduct(row) {
      const res = await findGeneralProduct({ ID: row.ID });
      this.type = "update";
      if (res.code == 0) {
        this.formData = res.data.regeneralProducts;
        this.dialogFormVisible = true;
      }
    },
    closeDialog() {
      this.dialogFormVisible = false;
      this.formData = {
          allowCancel:false,
          createAt:0,
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
          productPic1Middle:"",
          productPic1Small:"",
          productPic2Large:"",
          productPic2Middle:"",
          productPic2Small:"",
          productPic3Large:"",
          productPic3Middle:"",
          productPic3Small:"",
          productType:0,
          shortVideo:"",
          thumbnail:"",
          
      };
    },
    async deleteGeneralProduct(row) {
      this.visible = false;
      const res = await deleteGeneralProduct({ ID: row.ID });
      if (res.code == 0) {
        this.$message({
          type: "success",
          message: "删除成功"
        });
        if (this.tableData.length == 1) {
            this.page--;
        }
        this.getTableData();
      }
    },
    async enterDialog() {
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
        this.closeDialog();
        this.getTableData();
      }
    },
    openDialog() {
      this.type = "create";
      this.dialogFormVisible = true;
    }
  },
  async created() {
    await this.getTableData();
  
}
};
</script>

<style>
</style>