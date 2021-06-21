import React, { FC, useState, useRef } from 'react'
import { TextInput } from '../../src/components'
import { PencilAltIcon, SaveIcon, TrashIcon, XCircleIcon } from '@heroicons/react/outline'
import { Exact, SetTupleInput } from '../../src/generated/graphql'
import { UseMutationResponse } from 'urql'
import c from 'classnames'

type Props = {
	useSet: () => UseMutationResponse<any, Exact<{
		input: SetTupleInput
	}>>
	useUnset: () => UseMutationResponse<any, Exact<{
		input: SetTupleInput
	}>>
	finished: () => void
	keyName: string
	value: string
	appName: string
	isNew: boolean
}

export const TextInputPair: FC<Props> = ({ useSet, useUnset, finished, appName, keyName, value, isNew }) => {
	const [editing, setEditing] = useState<boolean>(isNew)
	const [loading, setLoading] = useState<boolean>(false)

	const keyRef = useRef(null)
	const valRef = useRef(null)

	const [setResult, set] = useSet()
	const [unsetResult, unset] = useUnset()

	const err = setResult.error || unsetResult.error

	const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()
		setLoading(true)

		const result = await set({
			input: {
				name:  appName,
				key:   keyRef.current.value,
				value: valRef.current.value,
			},
		})

		if (!result.error) {
			setEditing(false)
			setLoading(false)
			finished()
		}
	}

	const onDelete = async (key: string, value: string) => {
		setLoading(true)

		const result = await unset({
			input: {
				name:  appName,
				key:   key,
				value: value,
			},
		})

		if (!result.error) {
			setEditing(false)
			finished()
		}

		if (result.error) {
			setLoading(false)
		}
	}

	return 	(
		<>
			{err &&
			<div>{err.message}</div>}

			<form onSubmit={onSubmit} className={c('flex', { 'opacity-10': loading })}>
				<TextInput
					name="key"
					className="w-5/12"
					disabled={!isNew || loading}
					required
					ref={keyRef}
					defaultValue={keyName}
					placeholder="key"
				/>

				<TextInput
					name="value"
					className="w-5/12"
					disabled={(!editing && !isNew) || loading}
					required
					ref={valRef}
					defaultValue={value}
					placeholder="value"
				/>

				{editing
					? <XCircleIcon className="w-5 select-none" onClick={() => { setEditing(false); finished(); valRef.current.value = value }}/>
					: <PencilAltIcon className="w-5 select-none" onClick={() => setEditing(true)} />}

				{editing
					? <button type="submit"><SaveIcon className="w-5" /></button>
					: <TrashIcon className="w-5 text-red-600" onClick={() => { confirm(`Delete "${keyName}"?`) && onDelete(keyName, value) } }/>}
			</form>
		</>
	)
}