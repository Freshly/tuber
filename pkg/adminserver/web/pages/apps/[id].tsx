import { useRouter } from 'next/dist/client/router'
import React from 'react'
import { useGetFullAppQuery, useCreateReviewAppMutation } from '../../src/generated/graphql'
import { throwError } from '../../src/throwError'


const CreateForm = () => {
	const x = useCreateReviewAppMutation()

	console.log({ x })

	return <form>
		<input name="branchName" />
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
		<CreateForm />
	</div>
}

export default ShowApp
