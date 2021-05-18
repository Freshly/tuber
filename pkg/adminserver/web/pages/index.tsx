import { useQuery } from '@apollo/client'
import React from 'react'
import { GET_APPS } from '../src/operations/getApps'
import { GetApps } from '../src/operations/__generated__/GetApps'
import Link from 'next/link'

const HomePage = () => {
	const { loading, data, error } = useQuery<GetApps>(GET_APPS)

	if (error) {
		return <div>{error.message}</div>
	}

	return <>
		{loading
			? <div>loading</div>
			: data.getApps.map(app =>
				<div key={app.name}>
					<Link href={`/apps/${app.name}`}>{app.name}</Link>
				</div>,
			)}
	</>
}

export default HomePage
