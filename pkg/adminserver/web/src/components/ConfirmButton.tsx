import React, { useState } from 'react'

export const ConfirmButton = ({ title, useMutation, className, input }) => {
	const [{ error }, mutate] = useMutation()
	const [fetching, setFetching] = useState<boolean>(false)
	const [confirm, setConfirm] = useState<boolean>(false)
	const [success, setSuccess] = useState<boolean>(false)

	async function doMutation() {
		setConfirm(false)
		setFetching(true)
		const result = await mutate({ input: input })

		if (!result.error) {
			setSuccess(true)
			setTimeout(() => setSuccess(false), 1500)
		}

		setFetching(false)
	}

	return <div>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}

		{success
			? <div className="inline-block p-2 rounded-md border-2 border-light-800">Success!</div>
			: fetching
				? <div className="inline-block p-2 rounded-md border-2 border-light-800">Deploying...</div>
				: <>
					{confirm || <div className={`inline-block text-white p-2 rounded-md cursor-pointer border-2 ${className}`} onClick={() => setConfirm(true) }>
						{title}
					</div>}

					{confirm && <div className="inline-block text-green-700 bg-white p-2 rounded-md cursor-pointer border-2 border-green-700" onClick={() => doMutation() }>
					Confirm?
					</div>}

					{confirm && <div className="inline-block text-white bg-red-700 ml-2 p-2 rounded-md cursor-pointer border-2 border-red-700" onClick={() => setConfirm(false) }>
					Cancel
					</div>}
				</>
		}
	</div>
}
