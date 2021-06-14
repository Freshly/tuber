import React from 'react'
import Link from 'next/link'
import { useGetAppsQuery } from '../src/generated/graphql'
import { throwError } from '../src/throwError'
import { ClipboardCopyIcon } from '@heroicons/react/outline'

const HomePage = () => {
	const [{ data }] = throwError(useGetAppsQuery())

	return <section className="shadow-xl">
		{data.getApps.map(app =>
			<div key={app.name} className="border-b bg-white dark:bg-gray-800 block p-4 leading-4 flex justify-between">
				<div>
					<div className="pb-2 text-blue-500">
						<Link href={`/apps/${app.name}`} passHref>
							<a> {app.name} </a>
						</Link>
					</div>
					<div>
						<small className="text-light-900">Status: {app.paused ? 'Paused' : 'Running'} </small>
					</div>
				</div>

				<div>
					<small className="mr-1">{app.imageTag}</small>
					<ClipboardCopyIcon
						className="w-4 inline"
						onClick={() => { navigator.clipboard.writeText(app.imageTag) }}
					/>
				</div>
			</div>,
		)}
	</section>
}

export default HomePage
