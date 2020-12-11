///双色球基础数据结构
class ShuangSeQiu {
  ///胆拖区, 最多选5个红球, 从1-33号球中选取
  List<int> DantuoBalls;

  ///红球区, , 从1-33号球中选取。
  ///如果是胆拖，红球不能与胆拖已选的重复，只能从剩余里选 ，最多不超过20个
  ///如果是单式，红球只能是6个，并且每张单式彩票不能多于5注。
  ///如果是复式，红球必须大于或等于6个，不能超过33个
  List<int> RedBalls;

  ///篮球区, 从1-16号球中选取
  List<int> BlueBalls;
}

///双色球订单, 支持单式\复式\胆拖
class ShuangSeQiuOrder {
  ///买家用户注册账号, 必填
  String BuyUser;

  ///商户注册账号, 必填
  String BusinessUsername;

  ///商品ID, 必填
  String ProductID;

  ///订单商品ID, 支付成功后由服务端返回, 必填
  String OrderID;

  ///彩票投注类型，1-单式\2-复式\3-胆拖
  int TicketType = 0;

  //彩票的号码数组
  List<ShuangSeQiu> Straws;

  ///倍数 默认 1倍，单注不能超过50倍
  int Multiple = 1;

  ///总注数
  int Count = 0;

  ///总花费, 每注2.00元, 乘以倍数, 再乘以总注数
  double Cost = 0.00;

  ///拍照彩票照片的ObjID, 上传成功后，服务端返回，必填
  String LotteryPicObjID;

  //拍照彩票照片哈希, 上传成功后，服务端返回，必填
  String LotteryPicHash;
}

void main() {
  //例子1: 一注单式 6+1

  ShuangSeQiuOrder newOrder1 = new ShuangSeQiuOrder();
  newOrder1.BuyUser = 'id1';
  newOrder1.BusinessUsername = 'id3';
  newOrder1.ProductID = 'ba89b52c-eb97-4ce1-bb66-90b95cabffd1';
  newOrder1.OrderID = '2323sfsd-eb97-4ce1-2323-bb6623sdf322';
  newOrder1.TicketType = 1; //单式
  newOrder1.Straws = new List<ShuangSeQiu>();
  ShuangSeQiu straw = new ShuangSeQiu();
  straw.RedBalls = [1, 3, 18, 19, 21, 25]; //6个红球
  straw.BlueBalls = [9]; //1个篮球
  newOrder1.Straws.add(straw); //如果是多于一注，则再次增加
  newOrder1.Multiple = 1; //1倍
  newOrder1.Count = 1; //1注
  newOrder1.Cost = 2.00; //元
  newOrder1.LotteryPicObjID =
      'orders/id3/2020/12/10/xxxxx.jpg'; //先存放在网点用户id3的目录，等 id1确认后 就会移动到id1的目录
  newOrder1.LotteryPicHash = 'FEABC32123232BC1231323EED1123339898129382398CCCA';

  //例子2: 一注复式  7+2
  ShuangSeQiuOrder newOrder2 = new ShuangSeQiuOrder();
  newOrder2.BuyUser = 'id1';
  newOrder2.BusinessUsername = 'id3';
  newOrder2.ProductID = 'ba89b52c-eb97-4ce1-bb66-90b95cabffd1';
  newOrder2.OrderID = '2323sfsd-eb97-4ce1-2323-bb6623sdf322';
  newOrder2.TicketType = 2; //复式

  newOrder2.Straws = new List<ShuangSeQiu>();
  ShuangSeQiu straw2 = new ShuangSeQiu();
  straw2.RedBalls = [1, 3, 18, 19, 21, 25, 33]; //7个红球
  straw2.BlueBalls = [7, 13]; //2个篮球
  newOrder2.Straws.add(straw2);
  newOrder2.Multiple = 1; //1倍
  newOrder2.Count = 14; //14注
  newOrder2.Cost = 28.00; //元
  newOrder2.LotteryPicObjID =
      'orders/id3/2020/12/10/xxxxx.jpg'; //先存放在网点用户id3的目录，等 id1确认后 就会移动到id1的目录
  newOrder2.LotteryPicHash = 'FEABC32123232BC1231323EED1123339898129382398CCCA';

  //例子3: 一注胆拖，胆码 3 拖 5, 篮球 11

  ShuangSeQiuOrder newOrder3 = new ShuangSeQiuOrder();
  newOrder3.BuyUser = 'id1';
  newOrder3.BusinessUsername = 'id3';
  newOrder3.ProductID = 'ba89b52c-eb97-4ce1-bb66-90b95cabffd1';
  newOrder3.OrderID = '2323sfsd-eb97-4ce1-2323-bb6623sdf322';
  newOrder3.TicketType = 3; //胆拖
  newOrder3.Straws = new List<ShuangSeQiu>();

  ShuangSeQiu straw3 = new ShuangSeQiu();
  straw3.DantuoBalls = [1, 2, 5]; //胆码红球
  straw3.RedBalls = [3, 18, 19, 21, 33]; //拖码红球
  straw3.BlueBalls = [11]; //1个篮球
  newOrder3.Straws.add(straw3);

  newOrder3.Multiple = 1; //1倍
  newOrder3.Count = 10; //10注
  newOrder3.Cost = 20.00; //元
}
