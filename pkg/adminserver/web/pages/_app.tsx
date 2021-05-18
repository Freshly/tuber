import React from 'react'
import App from 'next/app'

import { ApolloClient, InMemoryCache } from '@apollo/client'
import { ApolloProvider } from '@apollo/client/react'

import 'windi.css'
import Link from 'next/link'

const client = new ApolloClient({
	uri:   'http://localhost:3001/tuber/graphql',
	cache: new InMemoryCache(),
})

const AppWrapper = props =>
	<ApolloProvider client={client}>
		<div className="p-3 bg-gray-800">
			<div className="container mx-auto">
				<h1><Link href="/"><a>Tuber Dashboard</a></Link></h1>
			</div>
		</div>

		<div className="container mx-auto py-3">
			<App {...props} />
		</div>
	</ApolloProvider>

export default AppWrapper
