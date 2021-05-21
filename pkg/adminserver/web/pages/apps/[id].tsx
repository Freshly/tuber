/* eslint-disable react/prop-types */
import app from 'next/app'
import { useRouter } from 'next/dist/client/router'
import React, { useRef } from 'react'
import { useGetFullAppQuery, useCreateReviewAppMutation } from '../../src/generated/graphql'
import { throwError } from '../../src/throwError'


const CreateForm = ({ app }) => {
	const [{ error }, create] = useCreateReviewAppMutation()
	const branchNameRef = useRef(null)

	const handle = (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()

		create({
			input: {
				name:       app.name,
				branchName: branchNameRef.current.value,
			},
		})
	}


	return <form onSubmit={handle}>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}
		<input name="branchName" ref={branchNameRef} />
		<button type="submit">Create</button>
	</form>
}

const ShowApp = () => {
	const router = useRouter()
	const id = router.query.id as string
	const [{ data: { getApp: app } }] = throwError(useGetFullAppQuery({ variables: { name: id } }))
	const hostname = `https://${app.name}.staging.freshlyservices.net/`

	return <div>
		<h1>{app.name}</h1>

		<p>
			Available at - <a href={hostname}>{hostname}</a> - if it uses your cluster&apos;s default hostname.
		</p>

		<h2>Create a review app</h2>
		<CreateForm app={app} />

		<h2>Review apps</h2>
		{app.reviewApps && app.reviewApps.map(reviewApp =>
			<div key={reviewApp.name}>{reviewApp.name}</div>,
		)}
	</div>
}

export default ShowApp
