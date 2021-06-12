import React, { FC, useState } from 'react'
import { SetTupleInput, Exact, Tuple } from '../generated/graphql'
import { UseMutationResponse } from 'urql'
import { TextInputPair } from './TextInputPair'
import { PlusCircleIcon } from '@heroicons/react/outline'

type Props = {
	vars: Pick<Tuple, 'key' | 'value'>[]
	appName: string
	useSet: () => UseMutationResponse<any, Exact<{ input: SetTupleInput }>>
	useUnset: () => UseMutationResponse<any, Exact<{ input: SetTupleInput }>>
}

export const TextInputGroup: FC<Props> = ({ vars, appName, useSet, useUnset }) => {
	const [addNew, setAddNew] = useState<boolean>(false)
	const finished = () => setAddNew(false)

	const compare = (a, b) => a.key > b.key ? 1 : -1

	return <>
		{vars.sort(compare).map(variable => 
			<TextInputPair
				useSet={useSet}
				useUnset={useUnset}
				finished={finished}
				key={variable.key}
				keyName={variable.key}
				value={variable.value}
				appName={appName}
				isNew={false}
			/>,
		)}

		{addNew
			? <TextInputPair
				useSet={useSet}
				useUnset={useUnset}
				finished={finished}
				keyName={''}
				value={''}
				appName={appName}
				isNew={true}
			/>
			: <button className="mt-3 flex align-middle" onClick={() => setAddNew(true)}><PlusCircleIcon className="w-5 mr-1 text-green-700 inline-block"/> Add New</button>}
	</>
}
