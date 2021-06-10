import React, { FC, useRef, useState } from 'react'
import { SetTupleInput, Exact, Tuple } from '../generated/graphql'
import { UseMutationResponse } from 'urql'
import { TextInput } from '../../src/components'
import { PencilAltIcon, SaveIcon, TrashIcon } from '@heroicons/react/outline'

type VarsFormProps = {
	vars: Pick<Tuple, 'key' | 'value'>[]
	appName: string
	setMutation: () => UseMutationResponse<any, Exact<{
		input: SetTupleInput
	}>>
	unsetMutation: () => UseMutationResponse<any, Exact<{
		input: SetTupleInput
	}>>
}

const VarsForm: FC<VarsFormProps> = ({ vars, appName, setMutation, unsetMutation }) => {
	const [editing, setEditing] = useState<boolean>(false)

	const keyRef = useRef(null)
	const valRef = useRef(null)
	
	const [setResult, set] = setMutation()
	const [unsetResult, unset] = unsetMutation()

	const submit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()

		const result = await set({
			input: {
				name:  appName,
				key:   keyRef.current.value,
				value: valRef.current.value,
			},
		})

		if (!result.error) {
			setEditing(false)
		}
	}

	return <>
		{setResult.error && <div className="bg-red-700 text-white border-red-700 p-2">
			{setResult.error.message}
		</div>}

		{unsetResult.error && <div className="bg-red-700 text-white border-red-700 p-2">
			{unsetResult.error.message}
		</div>}

		{vars.map(variable => 
			<form key={variable.key} onSubmit={submit} className="flex">
				<TextInput
					name="key"
					disabled={!editing}
					required
					ref={keyRef}
					defaultValue={variable.key}
					placeholder="key"
				/>

				<TextInput
					name="value"
					disabled={!editing}
					required
					ref={valRef}
					defaultValue={variable.value}
					placeholder="value"
				/>

				{editing
					? <button type="submit"><SaveIcon className="w-5" /></button>
					: <PencilAltIcon className="w-5" onClick={() => setEditing(true)} />}

				<TrashIcon className="w-5" onClick={() => unset({ input: { name: appName, key: variable.key, value: variable.value } })}/>
			</form>,
		)}
	</>
}

export default VarsForm

