import { PlusCircleIcon } from '@heroicons/react/outline'
import React from 'react'

export const AddButton = ({ onClick }) =>
	<button className="mt-3 flex align-middle" onClick={onClick}>
		<PlusCircleIcon className="w-5 mr-1 text-green-700 inline-block"/> Add New
	</button>