import { Hono } from "hono";
import { PrismaClient } from "prisma/generated/client";
import { errorResponse } from "@/utils/errorResponse";

const prisma = new PrismaClient();
const apod = new Hono();

apod.get("/", async (c) => {
	const latestApod = await prisma.pictures.findFirst({
		orderBy: {
			date: "desc",
		},
	});

	if (!latestApod) {
		return c.text("No APOD found", 404);
	}

	return c.json(latestApod);
});

apod.get("/:date", async (c) => {
	const date = c.req.param("date");
	const apodData = await prisma.pictures.findFirst({
		where: {
			date: date,
		},
	});
	if (!apodData) {
		return c.json(
			errorResponse("APOD not found", `No APOD found for date: ${date}`, 404),
			404,
		);
	}
	return c.json(apodData);
});

export default apod;
