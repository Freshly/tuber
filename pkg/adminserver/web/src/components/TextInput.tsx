/* eslint-disable react/prop-types */
import React, { FC, DetailedHTMLProps, InputHTMLAttributes } from 'react'

export const TextInput: FC<DetailedHTMLProps<InputHTMLAttributes<HTMLInputElement>, HTMLInputElement>> = React.forwardRef(({ className, ...rest }, ref) =>
	<input className={`dark:bg-gray-800 p-1 translate-x-6 mr-3 ${className}`} type="text" {...rest} ref={ref} />,
)

TextInput.displayName = 'TextInput'
