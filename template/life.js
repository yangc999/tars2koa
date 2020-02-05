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