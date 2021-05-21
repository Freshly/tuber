import React from 'react'
import Link from 'next/link'
import { useGetAppsQuery } from '../src/generated/graphql'
import { throwError } from '../src/throwError'

const HomePage = () => {
	const [{ data }] = throwError(useGetAppsQuery())

	return <>
		{data.getApps.map(app =>
			<div key={app.name}>
				<Link href={`/apps/${app.name}`}>{app.name}</Link>
			</div>,
		)}
	</>
}

export default HomePage
