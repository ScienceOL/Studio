import { NextRequest, NextResponse } from 'next/server';
import { config } from '@/configs/auth';

export async function GET(request: NextRequest) {
  try {
    // 从请求头中获取 Authorization
    const authorization = request.headers.get('authorization');

    if (!authorization) {
      return NextResponse.json(
        { code: 5012, error: { msg: 'Authorization header is required' } },
        { status: 401 }
      );
    }

    // 转发请求到后端 API
    const backendResponse = await fetch(`${config.apiBaseUrl}/api/user/info`, {
      method: 'GET',
      headers: {
        Authorization: authorization,
        'Content-Type': 'application/json',
      },
    });

    const data = await backendResponse.json();

    return NextResponse.json(data, {
      status: backendResponse.status,
    });
  } catch (error) {
    console.error('User info API error:', error);
    return NextResponse.json(
      { code: 5000, error: { msg: 'Internal server error' } },
      { status: 500 }
    );
  }
}
