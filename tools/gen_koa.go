package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

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

func (gen *GenKoa) genRouter() {
	gen.code.Reset()
	c := &gen.code
	c.WriteString("\n\"use strict\";\n\n")
	c.WriteString("var Stream = require(\"@tars/stream\");")
	c.WriteString("var Client = require(\"@tars/rpc\").client;\n")
	c.WriteString("var " + gen.p.Module + " = require(\"./" + gen.p.Source + "Proxy\")." + gen.p.Module + ";\n")
	c.WriteString("const router = require(\"koa-router\")();\n")
	for _, i := range gen.p.Interface {
		c.WriteString("var prx_" + i.TName + " = Client.stringToProxy(" + gen.p.Module + "." + i.TName + "Proxy, );\n")
		c.WriteString("router.post(\"/" + i.TName + "\", async (ctx, next) => {\n")
		c.WriteString("\t" + "try {\n")
		c.WriteString("\t\t" + "var tup_decode = new Stream.Tup();\n")
		c.WriteString("\t\t" + "tup_decode.decode(new Tars.BinBuffer(ctx.request.body));\n")
		c.WriteString("\t\t" + "var method = tup_decode.readString(\"method\");\n")
		c.WriteString("\t\t" + "swsitch (method) {\n")
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
				}
				c.WriteString("\t\t\t" + "var " + a.Name + ";\n")
			}
			c.WriteString("\t\t\t" + "let result = await prx_" + i.TName + "." + f.Name + "(" + inputArgs + ");\n")
			for _, a := range f.Args {
				if a.IsOut {
					c.WriteString("\t\t\t" + "result.response.arguments;\n")
				}
			}
			c.WriteString("\t\t\t" + "break;\n")
		}
		c.WriteString("\t\t" + "default:\n")
		c.WriteString("\t\t\t" + "console.log(\"error method:\", method);\n")
		c.WriteString("\t\t\t" + "throw \"err method\";\n")
		c.WriteString("\t\t" + "}\n")
		c.WriteString("\t" + "} catch (err) {\n")
		c.WriteString("\t\t" + "console.log(\"error:\" + err);\n")
		c.WriteString("\t\t" + "ctx.response.status = 500;\n")
		c.WriteString("\t\t" + "ctx.response.body = \"\";\n")
		c.WriteString("\t" + "}")
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
