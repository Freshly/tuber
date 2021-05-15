import { useQuery } from '@apollo/client'
import React from 'react'
import { GET_APPS } from '../src/operations/getApps'
import { GetApps } from '../src/operations/__generated__/GetApps'

const HomePage = () => {
	const { loading, data, error } = useQuery<GetApps>(GET_APPS)

	if (error) {
		return <div>{error.message}</div>
	}

	return <>
		<div>Welcome to Next.js!</div>

		{loading
			? <div>loading</div>
			: data.getApps.map(app =>
				<div key={app.name}>{app.name}</div>,
			)}
	</>
}

export default HomePage
