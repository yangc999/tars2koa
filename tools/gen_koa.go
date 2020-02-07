package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var gHost = flag.String("H", "127.0.0.1", "TarsServerHost")
var gPort = flag.String("P", "12000", "TarsServerPort")
var gThread = flag.String("T", "100", "TarsClientThread")
var gServerName = flag.String("N", "TestObj", "TarsServerObjName")

//GenKoa record go code information.
type GenKoa struct {
	code     bytes.Buffer
	vc       int      // var count. Used to generate unique variable names
	I        []string // imports with path
	path     string
	prefix   string
	tarsPath string
	p        *Parse
}

func (gen *GenKoa) genErr(err string) {
	panic(err)
}

func (gen *GenKoa) saveToSourceFile(filename string) {
	var beauty []byte
	var err error
	prefix := gen.prefix

	beauty = gen.code.Bytes()

	if filename == "stdout" {
		fmt.Println(string(beauty))
	} else {
		err = os.MkdirAll(prefix+gen.p.Module, 0766)

		if err != nil {
			gen.genErr(err.Error())
		}
		err = ioutil.WriteFile(prefix+gen.p.Module+"/"+filename, beauty, 0666)

		if err != nil {
			gen.genErr(err.Error())
		}
	}
}

func (gen *GenKoa) toTarsPath() string {
	ret := ""
	ret += *gServerName
	ret += "@"
	ret += "tcp"
	ret += " -h "
	ret += *gHost
	ret += " -p "
	ret += *gPort
	ret += " -t "
	ret += *gThread
	return ret
}

func (gen *GenKoa) genEureka() {
	gen.code.Reset()
	c := &gen.code
	c.WriteString(`
eureka:
  host: #eurekaIp
  port: #eurekaPort
  servicePath: "/eureka/apps/"
  
instance:
  app: #appName
  ipAddr: #runningIp
  port:
	"$": #runningPort
	"@enabled": "true"
  hostName: $(instance.ipAddr):$(instance.port.$)
  statusPageUrl: http://$(instance.hostName)/info
  healthCheckUrl: http://$(instance.hostName)/health
  vipAddress: "vip"
  dataCenterInfo:
	"@class": "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo"
	name: "MyOwn"	
	`)
	gen.saveToSourceFile("eureka.yml")
}

func (gen *GenKoa) genLife() {
	gen.code.Reset()
	c := &gen.code
	c.WriteString("\n\"use strict\";\n\n")
	c.WriteString(`
const life = require("koa-router")();
	
life.get("/info", async (ctx, next) => {
	ctx.set("Content-Type", "application/json");
	ctx.body = {};
});
	  
life.get("/health", async (ctx, next) => {
	ctx.set("Content-Type", "application/json");
	ctx.body = {status: "UP"};
});
	
module.exports = life
	`)
	gen.saveToSourceFile("life.js")
}

func (gen *GenKoa) toTypeName(ty *VarType) string {
	ret := ""
	switch ty.Type {
	case tkTBool:
		ret = "Stream.Boolean"
	case tkTInt:
		ret = "Stream.String"
	case tkTShort:
		ret = "Stream.String"
	case tkTByte:
		ret = "Stream.String"
	case tkTLong:
		ret = "Stream.Int64"
	case tkTFloat:
		ret = "Stream.Float"
	case tkTDouble:
		ret = "Stream.Double"
	case tkTString:
		ret = "Stream.String"
	case tkTVector:
		ret = "Stream.List"
		ret += "("
		ret += gen.toTypeName(ty.TypeK)
		ret += ")"
	case tkTMap:
		ret = "Stream.Map"
		ret += "("
		ret += gen.toTypeName(ty.TypeK)
		ret += ", "
		ret += gen.toTypeName(ty.TypeV)
		ret += ")"
	case tkName:
		vec := strings.Split(ty.TypeSt, "::")
		ret = gen.p.Module + "." + strings.Join(vec, ".")
	default:
		gen.genErr("Unknow Type " + TokenMap[ty.Type])
	}
	return ret
}

func (gen *GenKoa) toArgumentName(arg *ArgInfo) string {
	ret := ""
	switch arg.Type.Type {
	case tkTBool:
	case tkTInt:
	case tkTShort:
	case tkTByte:
	case tkTLong:
	case tkTFloat:
	case tkTDouble:
	case tkTString:
		ret = "\"" + arg.Name + "\""
	case tkTVector:
		ret = "\"" + arg.Name + "\""
		ret += ", "
		ret += gen.toTypeName(arg.Type)
	case tkTMap:
		ret = "\"" + arg.Name + "\""
		ret += ", "
		ret += gen.toTypeName(arg.Type)
	case tkName:
		ret = "\"" + arg.Name + "\""
		ret += ", "
		ret += gen.toTypeName(arg.Type)
	default:
		gen.genErr("Unknow Type " + TokenMap[arg.Type.Type])
	}
	return ret
}

func (gen *GenKoa) toFunctionName(ty *VarType) string {
	ret := ""
	switch ty.Type {
	case tkTBool:
		ret = "Boolean"
	case tkTInt:
		if ty.Unsigned {
			ret = "UInt32"
		} else {
			ret = "Int32"
		}
	case tkTShort:
		if ty.Unsigned {
			ret = "UInt16"
		} else {
			ret = "Int16"
		}
	case tkTByte:
		if ty.Unsigned {
			ret = "UInt8"
		} else {
			ret = "Int8"
		}
	case tkTLong:
		if ty.Unsigned {
			ret = "UInt64"
		} else {
			ret = "Int64"
		}
	case tkTFloat:
		ret = "Float"
	case tkTDouble:
		ret = "Double"
	case tkTString:
		ret = "String"
	case tkTVector:
		ret = "List"
	case tkTMap:
		ret = "Map"
	case tkName:
		ret = "Struct"
	default:
		gen.genErr("Unknow Type " + TokenMap[ty.Type])
	}
	return ret
}

func (gen *GenKoa) toReadFunctionName(arg *ArgInfo) string {
	tp := gen.toFunctionName(arg.Type)
	return "read" + tp + "(" + gen.toArgumentName(arg) + ")"
}

func (gen *GenKoa) toWriteFunctionName(arg *ArgInfo) string {
	tp := gen.toFunctionName(arg.Type)
	return "write" + tp + "(\"" + arg.Name + "\", " + arg.Name + ")"
}

func (gen *GenKoa) genRouter() {
	gen.code.Reset()
	c := &gen.code
	c.WriteString("\n\"use strict\";\n\n")
	c.WriteString("var Stream = require(\"@tars/stream\");")
	c.WriteString("var Client = require(\"@tars/rpc\").client;\n")
	head := strings.LastIndex(gen.p.Source, "/")
	tail := strings.LastIndex(gen.p.Source, ".tars")
	fileName := gen.p.Source[head+1 : tail]
	c.WriteString("var " + gen.p.Module + " = require(\"" + fileName + "Proxy\")." + gen.p.Module + ";\n")
	c.WriteString("const router = require(\"koa-router\")();\n")
	for _, i := range gen.p.Interface {
		c.WriteString("var prx_" + i.TName + " = Client.stringToProxy(" + gen.p.Module + "." + i.TName + "Proxy, \"" + gen.toTarsPath() + "\");\n")
		c.WriteString("router.post(\"/" + i.TName + "\", async (ctx, next) => {\n")
		c.WriteString("\t" + "try {\n")
		c.WriteString("\t\t" + "var tup_decode = new Stream.Tup();\n")
		c.WriteString("\t\t" + "var tup_encode = new Stream.Tup();\n")
		c.WriteString("\t\t" + "tup_decode.decode(new Tars.BinBuffer(ctx.request.body));\n")
		c.WriteString("\t\t" + "var method = tup_decode.readString(\"method\");\n")
		c.WriteString("\t\t" + "switch (method) {\n")
		for _, f := range i.Fun {
			c.WriteString("\t\t" + "case \"" + f.Name + "\":\n")
			inputArgs := ""
			for _, a := range f.Args {
				if !a.IsOut {
					if len(inputArgs) == 0 {
						inputArgs += a.Name
					} else {
						inputArgs += ", "
						inputArgs += a.Name
					}
					c.WriteString("\t\t\t" + "var " + a.Name + " = tup_decode." + gen.toReadFunctionName(&a) + ";\n")
				}
			}
			c.WriteString("\t\t\t" + "let result = await prx_" + i.TName + "." + f.Name + "(" + inputArgs + ");\n")
			for _, a := range f.Args {
				if a.IsOut {
					c.WriteString("\t\t\t" + "var " + a.Name + " = " + "result.response.arguments." + a.Name + ";\n")
					c.WriteString("\t\t\t" + "tup_encode." + gen.toWriteFunctionName(&a) + ";\n")
				}
			}
			c.WriteString("\t\t\t" + "break;\n")
		}
		c.WriteString("\t\t" + "default:\n")
		c.WriteString("\t\t\t" + "console.log(\"error method:\", method);\n")
		c.WriteString("\t\t\t" + "throw \"err method\";\n")
		c.WriteString("\t\t" + "}\n")
		c.WriteString("\t\t" + "var BinBuffer = tup_encode.encode(true);\n")
		c.WriteString("\t\t" + "ctx.response.body = BinBuffer.toNodeBuffer();\n")
		c.WriteString("\t" + "} catch (err) {\n")
		c.WriteString("\t\t" + "console.log(\"error:\" + err);\n")
		c.WriteString("\t\t" + "ctx.response.status = 500;\n")
		c.WriteString("\t" + "}\n")
		c.WriteString("});\n")
	}
	c.WriteString("module.exports = router")
	gen.saveToSourceFile("router.js")
}

func (gen *GenKoa) genKoa() {
	gen.code.Reset()
	c := &gen.code
	c.WriteString("\n\"use strict\";\n\n")
	c.WriteString(`
const Eureka = require("eureka-js-client").Eureka;
const client = new Eureka({
	filename: "eureka",
	cwd: "${__dirname}",
});
	
const app = require("koa")();
const life = require("life");
const router = require("router");
	
app.use(life.routes(), life.allowedMethods());
app.use(router.routes(), router.allowedMethods());
	
client.start(err => {
	app.listen(process.env.PORT || 3000, () => {
		console.log("server is running");
	});
});
	`)
	gen.saveToSourceFile("registry.js")
}

func (gen *GenKoa) genAll() {
	gen.genEureka()
	gen.genKoa()
	gen.genLife()
	gen.genRouter()
}

//Gen to parse file.
func (gen *GenKoa) Gen() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			// set exit code
			os.Exit(1)
		}
	}()

	gen.p = ParseFile(gen.path)
	gen.genAll()
}

//NewGenKoa build up a new path
func NewGenKoa(path string, outdir string) *GenKoa {
	if outdir != "" {
		b := []byte(outdir)
		last := b[len(b)-1:]
		if string(last) != "/" {
			outdir += "/"
		}
	}

	return &GenKoa{path: path, prefix: outdir}
}
