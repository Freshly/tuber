import React from 'react'
import App from 'next/app'
import { createClient, Provider } from 'urql'

import 'windi.css'
import Link from 'next/link'

const client = createClient({
	url:      'http://localhost:3001/tuber/graphql',
	suspense: true,
})

const AppWrapper = props =>
	<Provider value={client}>
		<div className="p-3 dark:bg-gray-800">
			<div className="container mx-auto">
				<h1><Link href="/"><a>Tuber Dashboard</a></Link></h1>
			</div>
		</div>

		<div className="container mx-auto py-3">
			<App {...props} />
		</div>
	</Provider>

export default AppWrapper