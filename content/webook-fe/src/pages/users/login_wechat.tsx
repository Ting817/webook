import React, { useState, useEffect } from 'react';
import {redirect} from "next/navigation";
import instance from "@/axios/axios";

function Page() {
    const [isLoading, setLoading] = useState(false)

    useEffect(() => {
        setLoading(true)
        instance.get('/oauth2/wechat/authurl')
            .then((res) => res.data)
            .then((data) => {
                setLoading(false)
                if(data && data.data) {
                    window.location.href = data.data
                }
            })
    }, [])

    if (isLoading) return <p>Loading...</p>

    return (
        <div>

        </div>
    )
}

export default Page