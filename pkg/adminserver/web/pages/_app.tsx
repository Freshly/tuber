import React, { Suspense } from 'react'
import App from 'next/app'
import { createClient, Provider } from 'urql'

import 'windi.css'
import { SiteHeader } from '../src/components'

const client = createClient({
	url:      `${process.env.TUBER_PREFIX}/graphql`,
	suspense: true,
})

const Loading = () => <div>loading...</div>

const AppWrapper = props =>
	<Provider value={client}>
		<SiteHeader />

		<div className="container mx-auto p-3">
			{typeof window !== 'undefined'
				? <Suspense fallback={<Loading />}>
					<App {...props} />
				</Suspense>
				: <Loading />}
		</div>
	</Provider>

export default AppWrapper
