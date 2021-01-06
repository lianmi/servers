
<template>
  <div>
    <div class="search-term">
      <el-form :inline="true" :model="searchInfo" class="demo-form-inline">                
        <el-form-item>
          <el-button @click="onSubmit" type="primary">查询</el-button>
        </el-form-item>
        <el-form-item>
          <el-button @click="openDialog" type="primary">新增lotterySaleTimes表</el-button>
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
    
    <el-table-column label="类型" prop="lotteryType" width="120"></el-table-column> 
    
    <el-table-column label="彩票名称" prop="lotteryName" width="120"></el-table-column> 
    
    <el-table-column label="开始的星期数" prop="saleStartWeek" width="120"></el-table-column> 
    
    <el-table-column label="开售时刻" prop="saleStartTime" width="120"></el-table-column> 
    
    <el-table-column label="停售星期数" prop="saleEndWeek" width="120"></el-table-column> 
    
    <el-table-column label="停售时刻 " prop="saleEndTime" width="120"></el-table-column> 
    
    <el-table-column label="节假日停售(选填)" prop="holidays" width="120"></el-table-column> 
    
    <el-table-column label="是否激活" prop="isActive" width="120">
         <template slot-scope="scope">{{scope.row.isActive|formatBoolean}}</template>
    </el-table-column>
    
      <el-table-column label="按钮组">
        <template slot-scope="scope">
          <el-button class="table-button" @click="updateLotterySaleTimes(scope.row)" size="small" type="primary" icon="el-icon-edit">变更</el-button>
          <el-popover placement="top" width="160" v-model="scope.row.visible">
            <p>确定要删除吗？</p>
            <div style="text-align: right; margin: 0">
              <el-button size="mini" type="text" @click="scope.row.visible = false">取消</el-button>
              <el-button type="primary" size="mini" @click="deleteLotterySaleTimes(scope.row)">确定</el-button>
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
    createLotterySaleTimes,
    deleteLotterySaleTimes,
    deleteLotterySaleTimesByIds,
    updateLotterySaleTimes,
    findLotterySaleTimes,
    getLotterySaleTimesList
} from "@/api/lotterySaleTimes";  //  此处请自行替换地址
import { formatTimeToStr } from "@/utils/date";
import infoList from "@/mixins/infoList";
export default {
  name: "LotterySaleTimes",
  mixins: [infoList],
  data() {
    return {
      listApi: getLotterySaleTimesList,
      dialogFormVisible: false,
      visible: false,
      type: "",
      deleteVisible: false,
      multipleSelection: [],formData: {
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
        if (this.searchInfo.isActive==""){
          this.searchInfo.isActive=null
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
        const res = await deleteLotterySaleTimesByIds({ ids })
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
    async updateLotterySaleTimes(row) {
      const res = await findLotterySaleTimes({ ID: row.ID });
      this.type = "update";
      if (res.code == 0) {
        this.formData = res.data.relotterySaleTimes;
        this.dialogFormVisible = true;
      }
    },
    closeDialog() {
      this.dialogFormVisible = false;
      this.formData = {
          lotteryType:0,
          lotteryName:"",
          saleStartWeek:"",
          saleStartTime:"",
          saleEndWeek:"",
          saleEndTime:"",
          holidays:"",
          isActive:false,
          
      };
    },
    async deleteLotterySaleTimes(row) {
      this.visible = false;
      const res = await deleteLotterySaleTimes({ ID: row.ID });
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