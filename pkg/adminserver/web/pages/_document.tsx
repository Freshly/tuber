import React from 'react'
import Document, { Html, Head, Main, NextScript } from 'next/document'

class MyDocument extends Document {
	static async getInitialProps(ctx) {
		const initialProps = await Document.getInitialProps(ctx)
		return { ...initialProps }
	}

	render() {
		return (
			<Html>
				<Head />
				<body className="font-sans text-gray-900 dark:text-gray-50 bg-gray-50 dark:bg-gray-900">
					<Main />
					<NextScript />
				</body>
			</Html>
		)
	}
}

export default MyDocument
