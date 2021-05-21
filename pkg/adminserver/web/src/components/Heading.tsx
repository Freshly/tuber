/* eslint-disable react/prop-types */
import React from 'react'

export const Heading = ({ children, size = '2', style = null }) => {
	if (!style) {
		switch (size) {
			case '1':
				style = '2xl'
				break
			default:
				style = 'lg'
				break
		}
	}

	return React.createElement(`h${size}`, { className: `pb-1 text-${style}` }, children)
}
