import { useRouter } from 'next/dist/client/router'
import React from 'react'
import { useGetAppQuery } from '../../src/generated/graphql'

const ShowApp = () => {
	const router = useRouter()
	const id = router.query.id as string
	const [{ data: { getApp: app } }] = useGetAppQuery({ variables: { name: id } })
	const hostname = `https://${app.name}.staging.freshlyservices.net/`

	return <div>
		<h1>{app.name}</h1>

		<p>
			Available at - <a href={hostname}>{hostname}</a> - if it uses your cluster&apos;s default hostname.
		</p>

		<h2>Create a review app</h2>
	</div>
}

export default ShowApp
