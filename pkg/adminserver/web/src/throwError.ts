import { UseQueryResponse } from 'urql'

export function throwError<T>(response: UseQueryResponse<T, object>): UseQueryResponse<T, object> {
	const [{ error }] = response

	if (error) {
		throw error
	}

	return response
}
