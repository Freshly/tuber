import { ChevronUpIcon, ChevronDownIcon } from '@heroicons/react/outline'
import React, { useState } from 'react'
import { TextInputGroup } from '.'
import { useQuery } from 'urql'

export const CollapsedAsyncTextInputGroup = ({ tableHeading, appName, queryDocument, queryVars, queryData, setMutation, unsetMutation }) => {
	const [expanded, setExpanded] = useState<boolean>(false)
	const [{ data, error }] = useQuery({ query: queryDocument, variables: queryVars, pause: !expanded })
	const appEnv = data ? queryData(data) : []
	return <div>
		<div className="flex justify-between mb-2" onClick={() => expanded ? setExpanded(false) : setExpanded(true) }>
			<h2 className="text-xl">{tableHeading}</h2>
			{expanded
				? <ChevronUpIcon className="w-6 relative"/>
				: <ChevronDownIcon className="w-6 relative"/>}
		</div>
		{error && <div>{error.message}</div>}
		{expanded && !error && <TextInputGroup
			appName={appName}
			vars={appEnv}
			useSet={setMutation}
			useUnset={unsetMutation}
		/>}
	</div>
}