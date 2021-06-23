import React, { Suspense } from 'react'
import Link from 'next/link'
import { useGetClusterInfoQuery } from '../generated/graphql'


const ClusterInfo = () => {
	const [{ data, error }] = useGetClusterInfoQuery()

	return error
		? <span>error</span>
		: <span className="text-sm ml-3 inline-block">{data.getClusterInfo.name}</span>
}

export const SiteHeader = () => {
	return <div className="bg-gray-100 dark:bg-gray-800">
		<div className="container mx-auto px-6 py-3">
			<h1 className="inline"><Link href="/"><a>Tuber Dashboard</a></Link></h1>

			{typeof window !== 'undefined' && <Suspense fallback={<span>loading</span>}>
				<ClusterInfo />
			</Suspense>}
		</div>
	</div>
}
