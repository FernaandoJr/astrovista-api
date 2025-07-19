interface HandlerLinks {
	totalPages: number;
	query?: string;
	startDate?: string;
	endDate?: string;
	mediaType?: string;
	perPage?: number;
	page?: number;
	sort?: string;
	hasNextPage?: boolean;
	hasPreviousPage?: boolean;
}

export function handlerLinks({
	hasNextPage,
	hasPreviousPage,
	totalPages,
	query,
	startDate,
	endDate,
	mediaType,
	perPage,
	page,
	sort,
}: HandlerLinks) {
	const baseUrl = "/apods/search?";

	const nextLink = hasNextPage
		? `${baseUrl}${new URLSearchParams({
				q: query || "",
				startDate: startDate || "",
				endDate: endDate || "",
				mediaType: mediaType || "",
				perPage: perPage ? String(perPage) : "",
				page: page ? String(page + 1) : "",
				sort: sort || "",
			})}`
		: null;

	const previousLink = hasPreviousPage
		? `${baseUrl}${new URLSearchParams({
				query: query || "",
				startDate: startDate || "",
				endDate: endDate || "",
				mediaType: mediaType || "",
				perPage: perPage ? String(perPage) : "",
				page: String((page ?? 1) - 1),
				sort: sort || "",
			})}`
		: null;

	const firstLink = `${baseUrl}${new URLSearchParams({
		q: query || "",
		startDate: startDate || "",
		endDate: endDate || "",
		mediaType: mediaType || "",
		perPage: perPage ? String(perPage) : "",
		page: "1",
		sort: sort || "",
	})}`;

	const lastLink = `${baseUrl}${new URLSearchParams({
		q: query || "",
		startDate: startDate || "",
		endDate: endDate || "",
		mediaType: mediaType || "",
		perPage: perPage ? String(perPage) : "",
		page: String(totalPages),
		sort: sort || "",
	})}`;

	const links = {
		next: nextLink,
		previous: previousLink,
		first: firstLink,
		last: lastLink === firstLink ? null : lastLink,
	};

	return links;
}
