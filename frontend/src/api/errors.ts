const serverMessage = "The server cannot complete the request. Try again in 1 minute.";
const rejectedMessage = "The server did not accept the request. Check the values. Then try again.";

// ApiError is the error of a request that the API did not accept. The message is a text for the person.
export class ApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

// detailOf gives the detail of a problem detail (RFC 9457), or undefined when the body has no text detail.
// The body can be anything: a proxy can send an empty body or an HTML page.
function detailOf(body: unknown): string | undefined {
  if (typeof body === "object" && body !== null && "detail" in body) {
    return typeof body.detail === "string" && body.detail !== "" ? body.detail : undefined;
  }
  return undefined;
}

// unwrap gives the data of a response that succeeded. Otherwise it throws an ApiError.
// It checks the status of the response, because a failed response can have no body.
export function unwrap<T>({
  data,
  error,
  response,
}: {
  data?: T;
  error?: unknown;
  response: Response;
}): T {
  if (response.ok && data !== undefined) {
    return data;
  }
  throw new ApiError(
    response.status,
    detailOf(error) ?? (response.status >= 500 ? serverMessage : rejectedMessage),
  );
}
