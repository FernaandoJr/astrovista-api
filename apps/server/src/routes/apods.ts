import { Hono } from "hono";
import { PrismaClient } from "prisma/generated/client";
import { errorResponse } from "@/utils/errorResponse";
import { handlerLinks } from "@/utils/handlerLinks";
import { searchResponse } from "@/utils/searchResponse";

const prisma = new PrismaClient();
const apods = new Hono();

apods.get("/", async (c) => {
	const apods = await prisma.pictures.findMany();
	return c.json(apods);
});

// ENDPOINT DE PESQUISA AVANÇADA
apods.get("/search", async (c) => {
	const query = c.req.query("q");
	const startDate = c.req.query("startDate");
	const endDate = c.req.query("endDate");
	const mediaType = c.req.query("mediaType");
	const perPage = c.req.query("perPage") ? Number(c.req.query("perPage")) : 10;
	const page = c.req.query("page") ? Number(c.req.query("page")) : 1;
	const sort = c.req.query("sort");

	if (
		(startDate &&
			endDate &&
			new Date(startDate as string) > new Date(endDate as string)) ||
		(startDate && isNaN(new Date(startDate as string).getTime())) ||
		(endDate && isNaN(new Date(endDate as string).getTime()))
	) {
		return c.json(
			errorResponse(
				"Invalid date range",
				"Start date cannot be after end date or invalid date format",
				400,
			),
			400,
		);
	}

	if (mediaType && !["image", "video"].includes(mediaType)) {
		return c.json(
			errorResponse(
				"Invalid media type",
				"Media type must be 'image' or 'video'",
				400,
			),
			400,
		);
	}

	if (perPage >= 200 || perPage < 1 || isNaN(perPage)) {
		return c.json(
			errorResponse(
				"Invalid perPage value",
				"perPage must be a number less than 200 and greater than 0",
				400,
			),
			400,
		);
	}
	if (page < 1) {
		return c.json(
			errorResponse("Invalid page number", "Page must be greater than 0", 400),
			400,
		);
	}

	if (sort && !["asc", "desc"].includes(sort)) {
		return c.json(
			errorResponse("Invalid sort value", "sort must be 'asc' or 'desc'", 400),
			400,
		);
	}

	const skip = page && perPage ? (page - 1) * perPage : undefined;

	const apods = await prisma.pictures.findMany({
		where: {
			title: {
				contains: query || "",
				mode: "insensitive",
			},
			date: {
				gte: startDate ? new Date(startDate).toISOString() : undefined,
				lte: endDate ? new Date(endDate).toISOString() : undefined,
			},
			media_type: {
				equals: mediaType || undefined,
			},
		},
		orderBy: {
			date: sort === "asc" ? "asc" : "desc",
		},
		take: perPage,
		skip: skip,
	});

	// count total records for pagination
	const totalRecords = await prisma.pictures.count({
		where: {
			title: {
				contains: query || "",
				mode: "insensitive",
			},
			date: {
				gte: startDate ? new Date(startDate).toISOString() : undefined,
				lte: endDate ? new Date(endDate).toISOString() : undefined,
			},
			media_type: {
				equals: mediaType || undefined,
			},
		},
		orderBy: {
			date: sort === "asc" ? "asc" : "desc",
		},
	});

	const hasNextPage = !!(page && perPage && apods.length === perPage);
	const hasPreviousPage = page > 1;

	if (apods.length === 0) {
		return c.json(
			errorResponse("No APODs found", "No results for the given query", 404),
			404,
		);
	}
	return c.json(
		searchResponse({
			apods,
			totalRecords,
			totalPages: Math.ceil(totalRecords / perPage),
			page: page,
			perPage: perPage,
			sort: sort || "desc",
			hasNextPage,
			hasPreviousPage,
			links: handlerLinks({
				hasNextPage,
				hasPreviousPage,
				totalPages: Math.ceil(totalRecords / perPage),
				query,
				startDate,
				endDate,
				mediaType,
				perPage,
				page,
				sort,
			}),
		}),
	);
});

export default apods;
