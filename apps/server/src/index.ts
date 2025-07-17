import { env } from "cloudflare:workers"
import { trpcServer } from "@hono/trpc-server"
import { createContext } from "./lib/context"
import { appRouter } from "./routers/index"
import { Hono } from "hono"
import { cors } from "hono/cors"
import { logger } from "hono/logger"
import { supabase } from "./db"

const app = new Hono()

app.use(logger())
app.use(
	"/*",
	cors({
		origin: env.CORS_ORIGIN || "",
		allowMethods: ["GET", "POST", "OPTIONS"],
	})
)

app.use(
	"/trpc/*",
	trpcServer({
		router: appRouter,
		createContext: (_opts, context) => {
			return createContext({ context })
		},
	})
)

app.get("/", (c) => {
	return c.text("OK")
})

app.get("/world", async (c) => {
	try {
		const { data, error } = await supabase
			.from("apods")
			.select("*")
			.limit(10)

		if (error) throw error

		return c.json({ apods: data })
	} catch (error) {
		console.error("Database error:", error)
		return c.json({ error: "Database connection failed" }, 500)
	}
})

export default app
