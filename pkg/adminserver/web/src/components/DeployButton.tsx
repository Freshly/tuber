import React, { useState } from 'react'
import { useDeployMutation } from '../../src/generated/graphql'

export const DeployButton = ({ appName }) => {
	const [{ error }, deploy] = useDeployMutation()
	const [fetching, setFetching] = useState<boolean>(false)
	const [confirm, setConfirm] = useState<boolean>(false)
	const [success, setSuccess] = useState<boolean>(false)

	async function doDeploy() {
		setConfirm(false)
		setFetching(true)
		const result = await deploy({ input: { name: appName } })

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
					{confirm || <div className="inline-block text-white bg-green-700 p-2 rounded-md cursor-pointer border-2 border-green-700" onClick={() => setConfirm(true) }>
					Deploy
					</div>}

					{confirm && <div className="inline-block text-green-700 bg-white p-2 rounded-md cursor-pointer border-2 border-green-700" onClick={() => doDeploy() }>
					Confirm?
					</div>}

					{confirm && <div className="inline-block text-white bg-red-700 ml-2 p-2 rounded-md cursor-pointer border-2 border-red-700" onClick={() => setConfirm(false) }>
					Cancel
					</div>}
				</>
		}
	</div>
}
