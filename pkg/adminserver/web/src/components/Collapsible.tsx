import React, { FC, useState } from 'react'

type Props = {
	collapsed: boolean
	children: React.ReactNode
}

export const Collapsible:FC<Props> = ({ collapsed = false, children }) => {
	const [expanded, setExpanded] = useState<boolean>(collapsed)

	return <div onClick={() => setExpanded(!expanded) } className={expanded ? 'expanded' : 'collapsed'}>
		{children}
	</div>
}