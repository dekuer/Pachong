/*
 * @Author: YeSiFan
 * @Date: 2019-04-10 15:24:41
 * @Last Modified by: YeSiFan
 * @Last Modified time: 2019-04-11 01:33:59
 */
package main

import "strconv"

/*
创建群结构体:属性包含群主、群昵称、群成员
*/

type Group struct {
	Name    string    //群昵称
	Owner   *Client   //群主
	Menbers []*Client //群成员
}

/*加群申请的回复*/
type GroupJoinReply struct {
	FromWhom *Client //回复的发送人
	ToWhom   *Client //申请人

	Group  *Group //申请的群
	Answer string //yes or no
}

/*建群工厂方法*/
func NewGroup(name string, owner *Client) *Group {
	group := new(Group)
	group.Name = name
	group.Owner = owner
	group.Menbers = make([]*Client, 0)
	group.Menbers = append(group.Menbers, owner)
	return group
}

/*
群昵称：xxxx
群主：xxxx
群人数：xxxx
*/

func (g *Group) String() string {
	info := "群昵称：" + g.Name + "\n"
	info += "群主：" + g.Owner.name + "\n"
	info += "群人数：" + strconv.Itoa(len(g.Menbers)) + "人\n"
	return info
}

/*添加新成员*/
func (g *Group) AddClient(client *Client) {

}

//申请加群回复
func NewGroupJoinReply(fromWhom, toWhom *Client, group *Group, answer string) *GroupJoinReply {
	reply := new(GroupJoinReply)
	reply.FromWhom = fromWhom
	reply.Answer = answer
	reply.Group = group
	reply.ToWhom = toWhom
	return reply
}

// 加群审核的执行申请
func (reply *GroupJoinReply) AutoRun() {
	// 是不是群主回复的
	if reply.FromWhom == reply.Group.Owner {
		// 如果是群主回复的
		if reply.Answer == "yes" {
			// 如果同意了
			reply.Group.AddClient(reply.ToWhom)
			SendMsg2Client("申请加入"+reply.Group.Name+"通过了", reply.ToWhom)
			SendMsg2Client(reply.ToWhom.name+"加入了"+reply.Group.Name, reply.Group.Owner)
		} else if reply.Answer == "no" {
			// 如果不同意
			SendMsg2Client("申请加入"+reply.Group.Name+"未通过,fuck off!", reply.ToWhom)
			SendMsg2Client(reply.ToWhom.name+"未加入"+reply.Group.Name, reply.Group.Owner)
		} else {
			SendMsg2Client("回复格式错了哦", reply.FromWhom)
		}
	} else {
		// 如果不是群主回复的则什么都不做
		SendMsg2Client("根据《中华人民治安反装逼法》，你已获得大牢三日游。", reply.FromWhom)
	}
}
