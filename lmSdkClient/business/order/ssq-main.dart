///双色球基础数据结构
///双色球注数及金额计算器: https://www.55128.cn/tool/ssq_tzmoney.aspx
class ShuangSeQiu {
  ///胆拖区, 最多选5个红球, 从1-33号球中选取
  List<int> DantuoBalls;

  ///红球区, , 从1-33号球中选取。
  ///如果是胆拖，红球不能与胆拖已选的重复，只能从剩余里选 ，胆码最多选5个, 最多不超过20个，篮球<=16个
  ///如果是单式，红球只能是6个，篮球1个，并且每张单式彩票不能多于5注。
  ///如果是复式，红球必须大于或等于6个，不能超过20个, 篮球必须大于或等于1个，<=16个。
  ///如果红球6个，篮球1个 ，不能视为复式
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

  ///订单商品ID, 预支付审核后由服务端返回, 必填
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

///复式注数计算器
int _calculateMultiple(int redBallsCount, int blueBallsCount) {
  if ((redBallsCount < 6) || (redBallsCount > 20)) {
    return 0;
  }

  if ((blueBallsCount < 1) || (blueBallsCount > 16)) {
    return 0;
  }

  if (redBallsCount == 6) {
    return blueBallsCount;
  }

  return _combin(redBallsCount, 6) * _combin(blueBallsCount, 1);
}

///胆拖注数计算器
int _calculateDantuo(int danmaCount, int tuomaCount, int blueBallsCount) {
  if ((danmaCount < 1) || (danmaCount > 5)) {
    return 0;
  }

  if (tuomaCount > 20) {
    return 0;
  }

  if ((blueBallsCount < 1) || (blueBallsCount > 16)) {
    return 0;
  }

  return _combin(tuomaCount, 6 - danmaCount) * _combin(blueBallsCount, 1);
}

int _factorial(int number) {
  var sum = 1;
  for (var i = 1; i <= number; i++) {
    sum *= i;
  }
  return sum;
}

int _combin(int n, m) {
  int a = _factorial(n);
  int b = _factorial(m);
  int n_m;
  if (n - m == 0) {
    n_m = 1;
  } else {
    n_m = _factorial(n - m);
  }

  int total = (a / (b * n_m)).round();

  return total;
}

void main() {
  //构造订单例子1: 一注单式 6+1
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

  ShuangSeQiu straw_2 = new ShuangSeQiu();
  straw_2.RedBalls = [2, 13, 14, 15, 22, 31]; //6个红球
  straw_2.BlueBalls = [7]; //1个篮球
  newOrder1.Straws.add(straw_2); //如果是多于一注，则再次增加

  newOrder1.Multiple = 1; //1倍
  newOrder1.Count = newOrder1.Straws.length; //2注
  newOrder1.Cost = newOrder1.Count * 2.00; //4元
  newOrder1.LotteryPicObjID =
      'orders/id3/2020/12/10/xxxxx.jpg'; //先存放在网点用户id3的目录，等 id1确认后 就会移动到id1的目录
  newOrder1.LotteryPicHash = 'FEABC32123232BC1231323EED1123339898129382398CCCA';
  print(
      "构造订单例子1: 2注单式6+1, 单倍, 总注数: ${newOrder1.Count}, 投注金额: ${newOrder1.Cost}");

  //构造订单例子2: 一注复式  7+2
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
  newOrder2.Count = _calculateMultiple(7, 2); //14注
  newOrder2.Cost = newOrder2.Count * 2.00; //28.0元
  newOrder2.LotteryPicObjID =
      'orders/id3/2020/12/10/xxxxx.jpg'; //先存放在网点用户id3的目录，等 id1确认后 就会移动到id1的目录
  newOrder2.LotteryPicHash = 'FEABC32123232BC1231323EED1123339898129382398CCCA';
  print(
      "构造订单例子2: 一注复式 7+2, 单倍, 总注数: ${newOrder2.Count}, 投注金额: ${newOrder2.Cost}");

  //构造订单例子3: 一注胆拖，胆码 3个 拖5个, 篮球 2个
  ShuangSeQiuOrder newOrder3 = new ShuangSeQiuOrder();
  newOrder3.BuyUser = 'id1';
  newOrder3.BusinessUsername = 'id3';
  newOrder3.ProductID = 'ba89b52c-eb97-4ce1-bb66-90b95cabffd1';
  newOrder3.OrderID = '2323sfsd-eb97-4ce1-2323-bb6623sdf322';
  newOrder3.TicketType = 3; //胆拖
  newOrder3.Straws = new List<ShuangSeQiu>();

  ShuangSeQiu straw3 = new ShuangSeQiu();
  straw3.DantuoBalls = [1, 2, 5]; //胆码红球3个
  straw3.RedBalls = [3, 18, 19, 21, 33]; //拖码红球5个
  straw3.BlueBalls = [11, 14]; //2个篮球
  newOrder3.Straws.add(straw3);

  newOrder3.Multiple = 1; //1倍
  newOrder3.Count = _calculateDantuo(3, 5, 2); //20注
  newOrder3.Cost = newOrder2.Count * 2.00; //40.0元
  print(
      "构造订单例子3: 一注胆拖，胆码 3个 拖5个, 篮球 2个, 单倍, 总注数: ${newOrder3.Count}, 投注金额: ${newOrder3.Cost}");

  print("==== 以下是更多复式例子");
  //复式例子: 9 + 2
  //结果验证 https://www.55128.cn/tool/ssq_tzmoney.aspx
  int total = _calculateMultiple(9, 2);
  double cost = total * 2.00;
  print("复式例子: 9 + 2, 总注数: ${total}, 投注金额: ${cost}");

  print("==== 以下是更多胆拖例子");
  //验证 https://www.78500.cn/tool/ssqdantuo.html

  //胆拖例子1: 红球区胆码个数2， 拖码个数是8， 篮球区个数是3
  int total2 = _calculateDantuo(2, 8, 3);
  cost = total2 * 2.00;
  print("胆拖例子1总注数: ${total2}, 投注金额: ${cost}");
}
