import React, { FC, useState } from 'react'
import { SetTupleInput, Exact, Tuple } from '../generated/graphql'
import { UseMutationResponse } from 'urql'
import { TextInputPair } from './TextInputPair'
import { AddButton } from './AddButton'

const compare = (a, b) => a.key > b.key ? 1 : -1

type Props = {
	vars: Pick<Tuple, 'key' | 'value'>[]
	appName: string
	useSet: () => UseMutationResponse<any, Exact<{ input: SetTupleInput }>>
	useUnset: () => UseMutationResponse<any, Exact<{ input: SetTupleInput }>>
}

export const TextInputGroup: FC<Props> = ({ vars, appName, useSet, useUnset }) => {
	const [addNew, setAddNew] = useState<boolean>(false)
	const finished = () => setAddNew(false)

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
			: <AddButton onClick={() => setAddNew(true)} />}
	</>
}
