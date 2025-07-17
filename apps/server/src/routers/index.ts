import { publicProcedure, router } from "../lib/trpc"

export const appRouter = router({
	healthCheck: publicProcedure.query(() => {
		return "OK"
	}),
	helloWorld: publicProcedure.query(() => {
		return { message: "Hello World!" }
	}),
})
export type AppRouter = typeof appRouter
