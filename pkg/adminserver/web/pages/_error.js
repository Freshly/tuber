/* eslint-disable react/prop-types */
import React from 'react'

function Error({ statusCode, message }) {

	return (
		<div className="bg-red-700 text-white border-red-700 border-4">
			<h2 className="bg-red-600 text-2xl p-3">Error: {statusCode || '500'}</h2>
			<h1>Maybe <a href="/tuber/login">log in?</a></h1>
			<div className="p-3">{message}</div>
		</div>
	)
}

Error.getInitialProps = ({ res, err }) => {
	const statusCode = res
		? res.statusCode
		: err
			? err.statusCode
			: 404

	const message = res === 404
		? 'Page not found'
		: err
			? err.message
			: 'An unexpected error occurred'

	return { statusCode, message }
}

export default Error
