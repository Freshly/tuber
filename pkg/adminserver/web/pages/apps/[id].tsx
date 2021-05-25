/* eslint-disable react/prop-types */
import { useRouter } from 'next/dist/client/router'
import React, { FC, useRef, useState } from 'react'
import { Heading } from '../../src/components'
import { TextInput } from '../../src/components/TextInput'
import { useGetFullAppQuery, useCreateReviewAppMutation, Tuple, useSetAppVarMutation } from '../../src/generated/graphql'
import { throwError } from '../../src/throwError'
import { PencilAltIcon, PlusCircleIcon, SaveIcon } from '@heroicons/react/outline'

const CreateForm = ({ app }) => {
	const [{ error }, create] = useCreateReviewAppMutation()
	const branchNameRef = useRef(null)

	const submit = (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()

		create({
			input: {
				name:       app.name,
				branchName: branchNameRef.current.value,
			},
		})
	}

	return <form onSubmit={submit}>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}
		<TextInput name="branchName" ref={branchNameRef} placeholder="branch name" />
		<button type="submit" className="rounded-sm p-1 underline">Create</button>
	</form>
}


type AppVarFormProps = {
	appVar: Tuple
	defaultEdit?: boolean
	name: string
	finished?: () => void
}

const AppVarForm: FC<AppVarFormProps> = ({ name, appVar, defaultEdit = false, finished }) => {
	const [editing, setEditing] = useState<boolean>(defaultEdit)
	const [{ error }, save] = useSetAppVarMutation()
	const keyRef = useRef(null)
	const valueRef = useRef(null)
	const formRef = useRef(null)

	const submit = async (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault()

		const result = await save({
			input: {
				name,
				key:   keyRef.current.value,
				value: valueRef.current.value,
			},
		})

		if (!result.error) {
			setEditing(false)
			finished && finished()
		}
	}

	return <form ref={formRef} onSubmit={submit}>
		{error && <div className="bg-red-700 text-white border-red-700 p-2">
			{error.message}
		</div>}

		<TextInput disabled={!editing || !defaultEdit} required name="key" ref={keyRef} defaultValue={appVar.key} placeholder="key" />
		<TextInput disabled={!editing} required name="value" ref={valueRef} defaultValue={appVar.value} placeholder="value" />

		{editing
			? <button type="submit"><SaveIcon className="w-5" /></button>
			: <PencilAltIcon className="w-5" onClick={() => setEditing(true)} />}
	</form>
}

const ShowApp = () => {
	const router = useRouter()
	const id = router.query.id as string
	const [{ data: { getApp: app } }] = throwError(useGetFullAppQuery({ variables: { name: id } }))
	const hostname = `https://${app.name}.staging.freshlyservices.net/`
	const [addNew, setAddNew] = useState<boolean>(false)

	return <div>
		<Heading>{app.name}</Heading>

		<p>
			Available at - <a href={hostname}>{hostname}</a> - if it uses your cluster&apos;s default hostname.
		</p>

		{app.reviewApp || <>
			<Heading>Create a review app</Heading>
			<CreateForm app={app} />

			<Heading>Review apps</Heading>
			{app.reviewApps && app.reviewApps.map(reviewApp =>
				<div key={reviewApp.name}>{reviewApp.name}</div>,
			)}

			<Heading>YAML Interpolation Vars</Heading>
			{app.vars.map(appVar => <AppVarForm key={appVar.key} name={app.name} appVar={appVar} />)}
			{addNew
				? <AppVarForm name={app.name} appVar={{} as Tuple} defaultEdit finished={() => setAddNew(false)} />
				: <PlusCircleIcon className="w-5" onClick={() => setAddNew(true)} />}
		</>}
	</div>
}

export default ShowApp
