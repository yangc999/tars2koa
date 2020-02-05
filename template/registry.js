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