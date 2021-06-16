import React from 'react'
import Link from 'next/link'
import { TuberApp, useGetAppsQuery } from '../src/generated/graphql'
import { throwError } from '../src/throwError'
import { Card, TextInput } from '../src/components'
import { useFuzzy } from 'react-use-fuzzy'

const HomePage = () => {
	const [{ data }] = throwError(useGetAppsQuery())
	const { result, search, keyword } = useFuzzy<Pick<TuberApp, 'name' | 'paused' | 'imageTag'>>(data.getApps, { keys: ['name', 'imageTag'] })

	return <>
		<TextInput
			placeholder="Search apps"
			value={keyword}
			className="mb-3 block w-[100%]"
			onChange={(e) => search(e.target.value)}
		/>

		<section className="grid grid-cols-3 gap-2 shadow-xl">
			{result.map(app =>
				<Card key={app.name}>
					<div className="flex align-middle justify-between">
						<Link href={`/apps/${app.name}`} passHref>
							<a className="text-blue-500 text-lg">
								<h1>{app.name}</h1>
							</a>
						</Link>

						<div>{app.paused && <small className="text-red-500">Paused</small>}</div>
					</div>

					<div>
						<small className="underline">
							<div><a href="https://console.cloud.google.com/" target="_blank" rel="noreferrer">GKE Workload</a></div>
							<div><a href="https://app.datadoghq.com/apm/home" target="_blank" rel="noreferrer">DataDog Logs</a></div>
						</small>
					</div>
				</Card>,
			)}
		</section>
	</>
}

export default HomePage
