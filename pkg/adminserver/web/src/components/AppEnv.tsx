import { ChevronUpIcon, ChevronDownIcon } from '@heroicons/react/outline'
import React, { FC, useState } from 'react'
import {
	useGetAppEnvQuery,
	useSetAppEnvMutation, 
	useUnsetAppEnvMutation,
} from '../generated/graphql'
import { TextInputGroup } from '../../src/components'
import Tuple from '../generated/graphql'

type Props = {
	collapsed?: boolean
	heading: string
	appName: string
}

export const AppEnv: FC<Props> = ({ collapsed = false, heading, appName }) => {
	const [expanded, setExpanded] = useState<boolean>(collapsed)
	var [{ data, error }] = useGetAppEnvQuery({ variables: { name: appName } })
	return <div>
		<div className="flex justify-between mb-2" onClick={() => setExpanded(!collapsed)}>
			<h2 className="text-xl">{heading}</h2>
			{expanded
				? <ChevronUpIcon className="w-6 relative"/>
				: <ChevronDownIcon className="w-6 relative"/>}
		</div>
		<div>
			{!expanded && <TextInputGroup
				appName={appName}
				vars={data.getAppEnv}
				useSet={useSetAppEnvMutation}
				useUnset={useUnsetAppEnvMutation}
			/>}
		</div>
	</div>
}

// function updateState(appName: string, expanded: boolean, setExpanded: React.Dispatch<React.SetStateAction<boolean>>) {
// 	await GetAppEnv(appName)
// 	if (expanded) {
// 		var [{data,error}] = useGetAppEnvQuery({ variables: { name: appName } })
// 		if (error) {
// 			return <div>aaaaaaaaaaaaaa</div>
// 		}
// 		return <TextInputGroup
// 			appName={appName}
// 			vars={data.getAppEnv}
// 			useSet={useSetAppEnvMutation}
// 			useUnset={useUnsetAppEnvMutation}
// 		/>
// 	} else {
// 		return <div></div>
// 	}
// 	setExpanded(expanded)
// }



// export async function GetAppEnv(appName: string) {
// 	const [{ data }] = useGetAppEnvQuery({ variables: { name: appName } })
// 	return <TextInputGroup
// 		appName={appName}
// 		vars={data.getAppEnv}
// 		useSet={useSetAppEnvMutation}
// 		useUnset={useUnsetAppEnvMutation}
// 	/>
// }

// function getAppEnv(appName: string) {
// 	const { data, error } = useSWR(useGetAppEnvQuery({ variables: { name: appName } }))

// 	return {
// 		data: data
// 	}
// }