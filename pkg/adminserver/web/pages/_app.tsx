import React, { Suspense } from 'react'
import App from 'next/app'
import { createClient, Provider } from 'urql'

import 'windi.css'
import Link from 'next/link'

const client = createClient({
	url:      `${process.env.TUBER_PREFIX}/graphql`,
	suspense: true,
})

const Loading = () => <div>loading...</div>

const AppWrapper = props =>
	<Provider value={client}>
		<div className="bg-gray-100 dark:bg-gray-800">
			<div className="container mx-auto p-3">
				<h1><Link href="/"><a>Tuber Dashboard</a></Link></h1>
			</div>
		</div>

		<div className="container mx-auto p-3">
			{typeof window !== 'undefined'
				? <Suspense fallback={<Loading />}>
					<App {...props} />
				</Suspense>
				: <Loading />}
		</div>
	</Provider>

export default AppWrapper
