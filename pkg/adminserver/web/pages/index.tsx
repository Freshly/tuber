import React from 'react'
import Link from 'next/link'
import { useGetAppsQuery } from '../src/generated/graphql'
import { throwError } from '../src/throwError'
import { ClipboardCopyIcon } from '@heroicons/react/outline'
import { Card } from '../src/components'

const HomePage = () => {
	const [{ data }] = throwError(useGetAppsQuery())

	return <section className="grid grid-cols-3 gap-2 shadow-xl">
		{data.getApps.map(app =>
			<Card key={app.name}>
				<Link href={`/apps/${app.name}`} passHref>
					<a className="text-blue-500 text-lg">
						<h1>{app.name}</h1>
					</a>vs
				</Link>

				<div>
					<small className="underline">
						<div><a href="https://console.cloud.google.com/" target="_blank" rel="noreferrer">GKE Workload</a></div>
						<div><a href="https://app.datadoghq.com/apm/home" target="_blank" rel="noreferrer">DataDog Logs</a></div>
					</small>

					{/* <small className="mr-1">{app.imageTag}</small>

						<ClipboardCopyIcon
							className="w-4 inline"
							onClick={() => { navigator.clipboard.writeText(app.imageTag) }}
						/> */}
				</div>

				<div>
					{app.paused && <h2 className="text-light-900">Status: Paused </h2>}
				</div>
			</Card>,
		)}
	</section>
}

export default HomePage
