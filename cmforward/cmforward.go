package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/mrspock/godocsis"
	"net"
	"os"
	"strings"
	//"strings"
)

const (
	VERSION string = "1.0.0"
	AUTHOR  string = "Marcin Jurczuk"
	EMAIL   string = "marcin@jurczuk.eu"
)

func main() {
	app := cli.NewApp()
	app.Name = "cmforward"
	app.Usage = "automatic add, remove forwarding rules for Technicolor TC7200 Cable modem and connected NP devices"
	app.Version = VERSION
	app.Author = AUTHOR
	app.Email = EMAIL
	app.CommandNotFound = func(ctx *cli.Context, command string) {
		fmt.Fprintf(os.Stderr, "\n%s action not found.\n\n", command)
		cli.ShowAppHelp(ctx)
	}
	app.Commands = []cli.Command{
		{
			Name:  "add",
			Usage: "add forwarding rules",
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "extport,ep",
					Value: 5001,
					Usage: "External starting port (incomming connection). If multiple devices are connected next will get export+1 value",
				},
			},
			Action: AddFwdRules,
		},
	}
	app.Run(os.Args)
}

func AddFwdRules(c *cli.Context) {
	var ip string
	var localIP string
	var startPort int
	startPort = c.Int("extport")
	fmt.Println("ExtPort:", startPort)
	var extIP string

	if len(c.Args()) < 1 {
		fmt.Println("Missing argument - cm ip address.")
		return
	} else {
		ip = c.Args().First()
		//localIP = flag.Arg(1)
	}

	s := godocsis.Session
	s.Target = ip
	s.Community = "private"
	// forward rules
	forwardRule := godocsis.CgForwardRule{}
	forwardRule.LocalIP = net.ParseIP(localIP)
	//forwardRule.RuleName = "Test"
	forwardRule.ExtPortStart = 5001
	forwardRule.LocalPortStart = 22
	forwardRule.ProtocolType = godocsis.Tcp
	forwardRule.IPAddrType = godocsis.IPv4
	//fmt.Println(s.Target, "device list:")
	devices, err := godocsis.CmGetNetiaPlayerList(s)
	if err != nil {
		fmt.Println("GetConnectedDevices:", err)
		return
	}
	if len(devices) > 0 {
		cm, err := godocsis.GetRouterIP(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: Unable to get externam router IP")
		}
		extIP = cm.RouterIP
	} else {
		fmt.Fprintf(os.Stderr, "No NetiaPlayer devices found on cable modem\n")
		os.Exit(0)
		extIP = "unknown"
	}
	for _, device := range devices {
		fmt.Println("NP detected:", device.IPAddr.String()+"\t"+device.MacAddr.String()+"\t"+device.Name)
		forwardRule.LocalIP = net.ParseIP(device.IPAddr.String())
		//forwardRule.ExtPortStart = startPort
		forwardRule.RuleName = strings.Replace(device.Name, "NetiaPlayer", "NP", -1)
		startPort++
		//fmt.Println(forwardRule)
		err = forwardRule.Validate()

		if err != nil {
			fmt.Println(err)
			continue
		}
		err = godocsis.CmSetForwardRule(s, &forwardRule, godocsis.TC7200ForwardingTree)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("OK:", s.Target, "rules set. CM Router IP: ", extIP, "external port: 5001")
	}
	//ruleCount, err := godocsis.CmGetFwdRuleCount(s, godocsis.TC7200ForwardingTree)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
	//fmt.Println("Liczba aktwnych reguł:", ruleCount)
	// fmt.Println(forwardRule)
}