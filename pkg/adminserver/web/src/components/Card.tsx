import React from 'react'

export const Card = ({ children, className = '' }) => {
	return <div className={`${className} bg-white dark:bg-gray-800 p-4 leading-4`}>
		{children}
	</div>
}
