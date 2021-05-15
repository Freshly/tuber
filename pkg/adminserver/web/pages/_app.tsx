import React from 'react'
import App from 'next/app'

import { ApolloClient, InMemoryCache } from '@apollo/client'
import { ApolloProvider } from '@apollo/client/react'

const client = new ApolloClient({
	uri:   'http://localhost:3001/tuber/graphql',
	cache: new InMemoryCache(),
})

const AppWrapper = props =>
	<ApolloProvider client={client}>
		<App {...props} />
	</ApolloProvider>

export default AppWrapper
