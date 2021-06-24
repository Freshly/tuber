import React from 'react'

export const Button = ({ title, useMutation, className, input }) => {
	const [{ error }, mutate] = useMutation()

	return <div>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}

		<div className={`inline-block text-white p-2 rounded-md cursor-pointer border-2 ${className}`} onClick={() => mutate({ input: input }) }>
			{title}
		</div>
	</div>
}
