import { pgTable, date, text } from "drizzle-orm/pg-core"

export const apods = pgTable("apods", {
	date: text("date").primaryKey(),
	explanation: text("explanation"),
	hdurl: text("hdurl"),
	media_type: text("media_type"),
	service_version: text("service_version"),
	title: text("title"),
	url: text("url"),
})
